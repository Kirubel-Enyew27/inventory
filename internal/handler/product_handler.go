package handler

import (
	"errors"
	"net/http"
	"strconv"

	"inventory/internal/model"
	"inventory/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

type createProductReq struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity" binding:"gte=0"`
	Price       float64 `json:"price" binding:"gt=0"`
}

type updateProductReq = createProductReq

func (h *ProductHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/products", h.CreateProduct)
	r.GET("/products", h.ListProducts)
	r.GET("/products/:id", h.GetProduct)
	r.PUT("/products/:id", h.UpdateProduct)
	r.DELETE("/products/:id", h.DeleteProduct)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req createProductReq
	if err := BindJSON(c, &req); err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	p := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	if err := h.svc.CreateProduct(c.Request.Context(), p); err != nil {
		if errors.Is(err, service.ErrInvalidPrice) || errors.Is(err, service.ErrInvalidQuantity) {
			JSONError(c, http.StatusBadRequest, err)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	category := c.Query("category")
	lowStock := false
	if c.Query("low_stock") == "true" {
		lowStock = true
	}

	// pagination
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			if v > 0 {
				if v > 100 {
					v = 100
				}
				limit = v
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	products, err := h.svc.GetAllProducts(c.Request.Context(), category, lowStock, limit, offset)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, products)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	p, err := h.svc.GetProductByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			JSONError(c, http.StatusNotFound, service.ErrNotFound)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}

	// ensure exists
	existing, err := h.svc.GetProductByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			JSONError(c, http.StatusNotFound, service.ErrNotFound)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}

	var req updateProductReq
	if err := BindJSON(c, &req); err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Category = req.Category
	existing.Quantity = req.Quantity
	existing.Price = req.Price

	if err := h.svc.UpdateProduct(c.Request.Context(), existing); err != nil {
		if errors.Is(err, service.ErrInvalidPrice) || errors.Is(err, service.ErrInvalidQuantity) {
			JSONError(c, http.StatusBadRequest, err)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid id"})
		return
	}
	if err := h.svc.DeleteProduct(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			JSONError(c, http.StatusNotFound, service.ErrNotFound)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusNoContent)
}
