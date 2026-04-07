package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	R "github.com/octopuswallet/octopuswallet/internal/api/response"
	"github.com/octopuswallet/octopuswallet/internal/api/errcode"
	"github.com/octopuswallet/octopuswallet/internal/store"
)

type ExportHandler struct {
	store store.Store
}

func NewExportHandler(s store.Store) *ExportHandler {
	return &ExportHandler{store: s}
}

func (h *ExportHandler) ExportPayments(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	from := c.DefaultQuery("from", "2020-01-01T00:00:00Z")
	to := c.DefaultQuery("to", "2099-12-31T23:59:59Z")
	format := c.DefaultQuery("format", "csv")

	payments, err := h.store.GetPaymentsByMerchantDateRange(c.Request.Context(), merchantID, from, to)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	if format == "json" {
		R.OK(c, gin.H{"payments": payments})
		return
	}

	// CSV
	var sb strings.Builder
	sb.WriteString("id,chain,token,amount_expected,amount_received,address,status,tx_hash,created_at\n")
	for _, p := range payments {
		txHash := ""
		if p.TxHash != nil {
			txHash = *p.TxHash
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			p.ID, p.Chain, p.Token, p.AmountExpected, p.AmountReceived,
			p.Address, p.Status, txHash, p.CreatedAt.Format("2006-01-02T15:04:05Z")))
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=payments.csv")
	c.String(http.StatusOK, sb.String())
}

func (h *ExportHandler) ExportPayouts(c *gin.Context) {
	merchantID := c.GetString("merchant_id")
	from := c.DefaultQuery("from", "2020-01-01T00:00:00Z")
	to := c.DefaultQuery("to", "2099-12-31T23:59:59Z")
	format := c.DefaultQuery("format", "csv")

	payouts, err := h.store.GetPayoutsByMerchantDateRange(c.Request.Context(), merchantID, from, to)
	if err != nil {
		R.Fail(c, errcode.ErrInternalServer)
		return
	}

	if format == "json" {
		R.OK(c, gin.H{"payouts": payouts})
		return
	}

	var sb strings.Builder
	sb.WriteString("id,chain,token,to_address,amount,status,approval_status,tx_hash,created_at\n")
	for _, p := range payouts {
		txHash := ""
		if p.TxHash != nil {
			txHash = *p.TxHash
		}
		sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			p.ID, p.Chain, p.Token, p.ToAddress, p.Amount,
			p.Status, p.ApprovalStatus, txHash, p.CreatedAt.Format("2006-01-02T15:04:05Z")))
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=payouts.csv")
	c.String(http.StatusOK, sb.String())
}
