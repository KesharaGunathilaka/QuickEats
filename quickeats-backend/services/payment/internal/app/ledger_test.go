package app

import (
	"testing"

	"github.com/quickeats/services/payment/internal/domain"
)

func TestPaymentEntriesCreditTheShopNotTheOrder(t *testing.T) {
	p := &domain.Payment{ID: "pay-1", OrderID: "order-1", ShopID: "shop-1", Amount: 45000}

	entries := paymentEntries(p, "cust-1")

	if got := entries[1].Account; got != "shop:shop-1" {
		t.Errorf("shop entry account = %s, want shop:shop-1", got)
	}
	for _, e := range entries {
		if e.Account == "shop:order-1" {
			t.Errorf("an entry still posts to the order ID: %+v", e)
		}
	}
}

func TestLedgerEntriesBalance(t *testing.T) {
	// 33333 doesn't split evenly 95/5: the parts must still add up to the payment.
	for _, amount := range []int64{45000, 12000, 33333, 1, 0} {
		p := &domain.Payment{ID: "pay-1", ShopID: "shop-1", Amount: amount}
		for name, entries := range map[string][]*domain.LedgerEntry{
			"payment": paymentEntries(p, "cust-1"),
			"refund":  refundEntries(p, "cust-1"),
		} {
			var debits, credits int64
			for _, e := range entries {
				if e.EntryType == domain.Debit {
					debits += e.Amount
				} else {
					credits += e.Amount
				}
			}
			if debits != credits || debits != amount {
				t.Errorf("%s of %d: debits %d, credits %d, want both %d", name, amount, debits, credits, amount)
			}
		}
	}
}

func TestRefundEntriesReversePayment(t *testing.T) {
	p := &domain.Payment{ID: "pay-1", ShopID: "shop-1", Amount: 45000}
	pay, refund := paymentEntries(p, "cust-1"), refundEntries(p, "cust-1")
	for i := range pay {
		if pay[i].Account != refund[i].Account || pay[i].Amount != refund[i].Amount || pay[i].EntryType == refund[i].EntryType {
			t.Errorf("entry %d: refund %+v doesn't reverse payment %+v", i, refund[i], pay[i])
		}
	}
}
