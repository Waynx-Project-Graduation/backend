package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kemit/trip-planner/internal/services"
	"github.com/kemit/trip-planner/internal/utils"
)

type MemberHandler struct {
	memberService *services.MemberService
}

func NewMemberHandler(memberService *services.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

// POST /api/trips/:id/members
func (h *MemberHandler) AddMember(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	var input services.AddMemberInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	member, err := h.memberService.AddMember(userID, tripID, input)
	if err != nil {
		switch err.Error() {
		case "trip not found", "user not found", "member not found":
			utils.NotFound(c, err.Error())
		case "only the trip owner can add members", "access denied":
			utils.Forbidden(c, err.Error())
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.Created(c, member)
}

// GET /api/trips/:id/members
func (h *MemberHandler) ListMembers(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}

	members, err := h.memberService.ListMembers(userID, tripID)
	if err != nil {
		if err.Error() == "access denied" {
			utils.Forbidden(c, err.Error())
			return
		}
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, members)
}

// PUT /api/trips/:id/members/:userId
func (h *MemberHandler) UpdateMemberRole(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}
	memberUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		utils.BadRequest(c, "invalid user ID")
		return
	}

	var input services.UpdateMemberRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := h.memberService.UpdateMemberRole(userID, tripID, memberUserID, input); err != nil {
		switch err.Error() {
		case "trip not found", "member not found":
			utils.NotFound(c, err.Error())
		case "only the trip owner can update member roles":
			utils.Forbidden(c, err.Error())
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.Success(c, gin.H{"message": "member role updated successfully"})
}

// DELETE /api/trips/:id/members/:userId
func (h *MemberHandler) RemoveMember(c *gin.Context) {
	userID := getUserID(c)
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "invalid trip ID")
		return
	}
	memberUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		utils.BadRequest(c, "invalid user ID")
		return
	}

	if err := h.memberService.RemoveMember(userID, tripID, memberUserID); err != nil {
		switch err.Error() {
		case "trip not found", "member not found":
			utils.NotFound(c, err.Error())
		case "only the trip owner can remove members":
			utils.Forbidden(c, err.Error())
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	utils.Success(c, gin.H{"message": "member removed successfully"})
}
