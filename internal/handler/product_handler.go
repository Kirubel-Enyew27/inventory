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
	SKU         string  `json:"sku" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity" binding:"gte=0"`
	Price       float64 `json:"price" binding:"gt=0"`
}

type updateProductReq = createProductReq

type adjustStockReq struct {
	Delta *int `json:"delta" binding:"required"`
}

type listProductsResp struct {
	Data   []model.Product `json:"data"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Total  int64           `json:"total"`
}

func (h *ProductHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/products", h.CreateProduct)
	r.GET("/products", h.ListProducts)
	r.GET("/products/:id", h.GetProduct)
	r.PUT("/products/:id", h.UpdateProduct)
	r.PATCH("/products/:id/stock", h.AdjustStock)
	r.DELETE("/products/:id", h.DeleteProduct)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req createProductReq
	if err := BindJSON(c, &req); err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	p := &model.Product{
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Price:       req.Price,
	}

	if err := h.svc.CreateProduct(c.Request.Context(), p); err != nil {
		if errors.Is(err, service.ErrInvalidSKU) || errors.Is(err, service.ErrInvalidPrice) || errors.Is(err, service.ErrInvalidQuantity) {
			JSONError(c, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, service.ErrDuplicateSKU) {
			JSONError(c, http.StatusConflict, service.ErrDuplicateSKU)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("q")
	sku := c.Query("sku")
	lowStock, err := parseBoolQuery(c, "low_stock", false)
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	limit, err := parseIntQuery(c, "limit", 20, 1, 100)
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}
	offset, err := parseIntQuery(c, "offset", 0, 0, 0)
	if err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}

	products, total, err := h.svc.GetAllProducts(c.Request.Context(), category, lowStock, search, sku, limit, offset)
	if err != nil {
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, listProductsResp{
		Data:   products,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "invalid id")
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
		JSONError(c, http.StatusBadRequest, "invalid id")
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
	existing.SKU = req.SKU
	existing.Description = req.Description
	existing.Category = req.Category
	existing.Quantity = req.Quantity
	existing.Price = req.Price

	if err := h.svc.UpdateProduct(c.Request.Context(), existing); err != nil {
		if errors.Is(err, service.ErrInvalidSKU) || errors.Is(err, service.ErrInvalidPrice) || errors.Is(err, service.ErrInvalidQuantity) {
			JSONError(c, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, service.ErrDuplicateSKU) {
			JSONError(c, http.StatusConflict, service.ErrDuplicateSKU)
			return
		}
		JSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, existing)
}

func (h *ProductHandler) AdjustStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "invalid id")
		return
	}

	var req adjustStockReq
	if err := BindJSON(c, &req); err != nil {
		JSONError(c, http.StatusBadRequest, err)
		return
	}
	if *req.Delta == 0 {
		JSONError(c, http.StatusBadRequest, "delta cannot be zero")
		return
	}

	product, err := h.svc.AdjustStock(c.Request.Context(), uint(id), *req.Delta)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			JSONError(c, http.StatusNotFound, service.ErrNotFound)
		case errors.Is(err, service.ErrInsufficientStock):
			JSONError(c, http.StatusConflict, service.ErrInsufficientStock)
		default:
			JSONError(c, http.StatusInternalServerError, err)
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "invalid id")
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

func parseBoolQuery(c *gin.Context, name string, fallback bool) (bool, error) {
	raw := c.Query(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, errors.New(name + " must be a boolean")
	}
	return value, nil
}

func parseIntQuery(c *gin.Context, name string, fallback, min, max int) (int, error) {
	raw := c.Query(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.New(name + " must be an integer")
	}
	if value < min {
		return 0, errors.New(name + " is below the minimum")
	}
	if max > 0 && value > max {
		return 0, errors.New(name + " is above the maximum")
	}
	return value, nil
}
