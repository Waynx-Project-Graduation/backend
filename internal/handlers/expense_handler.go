package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type ExpenseHandler struct {
	expenseService *services.ExpenseService
}

func NewExpenseHandler(expenseService *services.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{expenseService: expenseService}
}

// POST /api/trips/:id/expenses
func (h *ExpenseHandler) CreateExpense(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	var input services.CreateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	expense, err := h.expenseService.CreateExpense(userID, tripID, input)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		if err.Error() == "trip not found" {
			utils.NotFound(c, err.Error())
			return
		}
		utils.BadRequest(c, err.Error())
		return
	}

	utils.Created(c, expense)
}

// GET /api/trips/:id/expenses
func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	expenses, total, err := h.expenseService.ListExpenses(userID, tripID, page, perPage)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.SuccessWithMeta(c, expenses, &utils.Meta{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	})
}

// PUT /api/trips/:id/expenses/:expenseId
func (h *ExpenseHandler) UpdateExpense(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}
	expenseID, err := uuid.Parse(c.Param("expenseId"))
	if err != nil {
		utils.BadRequest(c, "invalid expense ID")
		return
	}

	var input services.UpdateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	expense, err := h.expenseService.UpdateExpense(userID, tripID, expenseID, input)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, expense)
}

// DELETE /api/trips/:id/expenses/:expenseId
func (h *ExpenseHandler) DeleteExpense(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}
	expenseID, err := uuid.Parse(c.Param("expenseId"))
	if err != nil {
		utils.BadRequest(c, "invalid expense ID")
		return
	}

	if err := h.expenseService.DeleteExpense(userID, tripID, expenseID); err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "expense deleted successfully"})
}
