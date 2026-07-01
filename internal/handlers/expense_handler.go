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

// CreateExpense godoc
// @Summary      Create expense
// @Description  Add a new expense to a trip
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        input body services.CreateExpenseInput true "Expense info"
// @Success      201  {object}  models.TripExpense
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/expenses [post]
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

// ListExpenses godoc
// @Summary      List expenses
// @Description  Get a paginated list of expenses for a trip
// @Tags         expenses
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        page query int false "Page number" default(1)
// @Param        per_page query int false "Items per page" default(10)
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/expenses [get]
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

// UpdateExpense godoc
// @Summary      Update expense
// @Description  Update an existing expense
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        expenseId path string true "Expense ID (UUID)"
// @Param        input body services.UpdateExpenseInput true "Expense update info"
// @Success      200  {object}  models.TripExpense
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/expenses/{expenseId} [put]
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

// DeleteExpense godoc
// @Summary      Delete expense
// @Description  Delete an expense from a trip
// @Tags         expenses
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        expenseId path string true "Expense ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/expenses/{expenseId} [delete]
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
