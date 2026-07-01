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

// AddMember godoc
// @Summary      Add member
// @Description  Add a new member to a trip
// @Tags         members
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        input body services.AddMemberInput true "Member info"
// @Success      201  {object}  models.TripMember
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/members [post]
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

// ListMembers godoc
// @Summary      List members
// @Description  Get a list of members for a trip
// @Tags         members
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Success      200  {array}   models.TripMember
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/members [get]
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

// UpdateMemberRole godoc
// @Summary      Update member role
// @Description  Update the role of a trip member
// @Tags         members
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        userId path string true "User ID (UUID)"
// @Param        input body services.UpdateMemberRoleInput true "Role update info"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/members/{userId} [put]
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

// RemoveMember godoc
// @Summary      Remove member
// @Description  Remove a member from a trip
// @Tags         members
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Trip ID (UUID)"
// @Param        userId path string true "User ID (UUID)"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /trips/{id}/members/{userId} [delete]
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
