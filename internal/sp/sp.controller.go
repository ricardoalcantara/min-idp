package sp

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	dbpkg "github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	sp_dto "github.com/ricardoalcantara/min-idp/internal/sp/dto"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
	"github.com/ricardoalcantara/min-idp/internal/users"
)

type SPController struct {
	service *SPService
	rbacSvc *rbac.RBACService
	userSvc *users.UserService
}

func NewSPController(service *SPService, rbacSvc *rbac.RBACService, userSvc *users.UserService) *SPController {
	return &SPController{service: service, rbacSvc: rbacSvc, userSvc: userSvc}
}

func (c *SPController) list(ctx *gin.Context) {
	sps, err := c.service.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]sp_dto.SPDto, len(sps))
	for i := range sps {
		dtos[i] = sp_dto.NewSPDto(&sps[i])
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *SPController) create(ctx *gin.Context) {
	var input sp_dto.CreateSPDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	sp, err := c.service.Create(input.Slug, input.Name, input.Protocol)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusCreated, sp_dto.NewSPDto(sp))
}

func (c *SPController) get(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.NewSPDto(sp))
}

func (c *SPController) update(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input sp_dto.UpdateSPDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	updated, err := c.service.Update(sp, input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.NewSPDto(updated))
}

func (c *SPController) delete(ctx *gin.Context) {
	if err := c.service.Delete(ctx.Param("id")); err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *SPController) getOIDC(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	client, err := c.service.GetOIDCClient(sp)
	if err != nil {
		if errors.Is(err, errProtocolMismatch) {
			ctx.JSON(http.StatusUnprocessableEntity, web.NewErrorDto(err))
			return
		}
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(errors.New("oidc client not configured")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.OIDCClientDto{
		ClientID:          client.ClientID,
		RedirectURIs:      sp_repositories.UnmarshalStringSlice(client.RedirectURIs),
		GrantTypes:        sp_repositories.UnmarshalStringSlice(client.GrantTypes),
		ResponseTypes:     sp_repositories.UnmarshalStringSlice(client.ResponseTypes),
		Scopes:            sp_repositories.UnmarshalStringSlice(client.Scopes),
		TokenEndpointAuth: client.TokenEndpointAuth,
		PKCERequired:      client.PKCERequired,
	})
}

func (c *SPController) putOIDC(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input sp_dto.UpsertOIDCClientDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	client, err := c.service.UpsertOIDCClient(sp, input)
	if err != nil {
		if errors.Is(err, errProtocolMismatch) {
			ctx.JSON(http.StatusUnprocessableEntity, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.OIDCClientDto{
		ClientID:          client.ClientID,
		RedirectURIs:      sp_repositories.UnmarshalStringSlice(client.RedirectURIs),
		GrantTypes:        sp_repositories.UnmarshalStringSlice(client.GrantTypes),
		ResponseTypes:     sp_repositories.UnmarshalStringSlice(client.ResponseTypes),
		Scopes:            sp_repositories.UnmarshalStringSlice(client.Scopes),
		TokenEndpointAuth: client.TokenEndpointAuth,
		PKCERequired:      client.PKCERequired,
	})
}

func (c *SPController) getSAML(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	client, err := c.service.GetSAMLClient(sp)
	if err != nil {
		if errors.Is(err, errProtocolMismatch) {
			ctx.JSON(http.StatusUnprocessableEntity, web.NewErrorDto(err))
			return
		}
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(errors.New("saml client not configured")))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.SAMLClientDto{
		EntityID:             client.EntityID,
		ACSURLs:              sp_repositories.UnmarshalStringSlice(client.ACSURLs),
		SLOUrl:               client.SLOUrl,
		NameIDFormat:         client.NameIDFormat,
		WantSignedRequests:   client.WantSignedRequests,
		WantSignedAssertions: client.WantSignedAssertions,
	})
}

func (c *SPController) putSAML(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input sp_dto.UpsertSAMLClientDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	client, err := c.service.UpsertSAMLClient(sp, input)
	if err != nil {
		if errors.Is(err, errProtocolMismatch) {
			ctx.JSON(http.StatusUnprocessableEntity, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusOK, sp_dto.SAMLClientDto{
		EntityID:             client.EntityID,
		ACSURLs:              sp_repositories.UnmarshalStringSlice(client.ACSURLs),
		SLOUrl:               client.SLOUrl,
		NameIDFormat:         client.NameIDFormat,
		WantSignedRequests:   client.WantSignedRequests,
		WantSignedAssertions: client.WantSignedAssertions,
	})
}

func (c *SPController) listRules(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	rows, err := c.service.ListAccessRules(sp.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]sp_dto.AccessRuleDto, len(rows))
	for i, row := range rows {
		dtos[i] = sp_dto.AccessRuleDto{
			ID:          row.AccessRule.UUID.String(),
			RuleType:    row.RuleType,
			SubjectType: row.SubjectType,
			SubjectID:   row.SubjectUUID,
			Priority:    row.Priority,
		}
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *SPController) createRule(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	var input sp_dto.CreateAccessRuleDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}

	entityID, err := c.resolveSubjectEntityID(ctx, input.SubjectType, input.SubjectID)
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	rule, err := c.service.CreateAccessRule(sp, input.SubjectType, entityID, input.RuleType, input.Priority)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.JSON(http.StatusCreated, sp_dto.AccessRuleDto{
		ID:          rule.UUID.String(),
		RuleType:    rule.RuleType,
		SubjectType: input.SubjectType,
		SubjectID:   input.SubjectID,
		Priority:    rule.Priority,
	})
}

func (c *SPController) deleteRule(ctx *gin.Context) {
	sp, err := c.service.FindByUUID(ctx.Param("id"))
	if err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if err := c.service.DeleteAccessRule(sp, ctx.Param("ruleId")); err != nil {
		if errors.Is(err, dbpkg.ErrEntityNotFound) {
			ctx.JSON(http.StatusNotFound, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Status(http.StatusNoContent)
}

// resolveSubjectEntityID converts a subject UUID to its DB entity ID.
func (c *SPController) resolveSubjectEntityID(ctx *gin.Context, subjectType, subjectUUID string) (uint, error) {
	switch subjectType {
	case dbpkg.SubjectTypeRole:
		role, err := c.rbacSvc.FindRoleByUUID(subjectUUID)
		if err != nil {
			return 0, err
		}
		return role.ID, nil
	case dbpkg.SubjectTypeUser:
		u, err := c.userSvc.FindByUUID(subjectUUID)
		if err != nil {
			return 0, err
		}
		return u.ID, nil
	default:
		return 0, errors.New("unknown subject type")
	}
}
