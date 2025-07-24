package delivery

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "video-call/internal/models"
    "video-call/internal/signaling"
)

type restDelivery struct {
    usecase signaling.CallUsecase
}

func NewRESTDelivery(usecase signaling.CallUsecase) signaling.CallRESTDelivery {
    return &restDelivery{usecase: usecase}
}

type callRequest struct {
    CallerID string `json:"caller_id" binding:"required,uuid"`
    CalleeID string `json:"callee_id" binding:"required,uuid"`
}

type callResponse struct {
    Call *models.Call `json:"call"`
    Role string       `json:"role"`
}

func (d *restDelivery) RegisterRoutes(rg *gin.RouterGroup) {
    rg.POST("/call", d.createOrJoinCall)
}

func (d *restDelivery) createOrJoinCall(c *gin.Context) {
    var req callRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    callerID, err := uuid.Parse(req.CallerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid caller_id"})
        return
    }
    calleeID, err := uuid.Parse(req.CalleeID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid callee_id"})
        return
    }
    call, role, err := d.usecase.CreateOrJoinCall(c.Request.Context(), callerID, calleeID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, callResponse{Call: call, Role: role})
} 