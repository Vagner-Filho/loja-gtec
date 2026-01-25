package checkout

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"lojagtec/internal/orders"

	"github.com/stripe/stripe-go/v84"
	checkoutsession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

var ErrStripeNotConfigured = errors.New("stripe_not_configured")

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func CreateCheckoutSession(form orders.CheckoutForm, order *orders.Order) (string, error) {
	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey == "" {
		return "", ErrStripeNotConfigured
	}

	stripe.Key = stripeKey

	paymentMethodTypes, err := stripePaymentMethodTypes(form.PaymentMethod)
	if err != nil {
		return "", err
	}

	lineItems, err := stripeLineItems(form.CartItems)
	if err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	params := &stripe.CheckoutSessionParams{
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: paymentMethodTypes,
		LineItems:          lineItems,
		SuccessURL:         stripe.String(fmt.Sprintf("%s/checkout/success?order_id=%d&session_id={CHECKOUT_SESSION_ID}", baseURL, order.ID)),
		CancelURL:          stripe.String(fmt.Sprintf("%s/checkout/cancel?order_id=%d", baseURL, order.ID)),
		CustomerEmail:      stripe.String(form.Email),
		ClientReferenceID:  stripe.String(strconv.Itoa(order.ID)),
		Metadata: map[string]string{
			"order_id":       strconv.Itoa(order.ID),
			"order_number":   order.OrderNumber,
			"cpf_cnpj":       form.CPF,
			"payment_method": form.PaymentMethod,
		},
	}

	stripeSession, err := checkoutsession.New(params)
	if err != nil {
		return "", err
	}

	if err := orders.UpdateOrderStripePaymentID(order.ID, stripeSession.ID); err != nil {
		return "", err
	}

	return stripeSession.URL, nil
}

func HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	webhookSecret := strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if webhookSecret == "" {
		http.Error(w, "Webhook secret not configured", http.StatusInternalServerError)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("read payload: %v", err.Error())
		http.Error(w, "Failed to read payload", http.StatusBadRequest)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sigHeader, webhookSecret)
	if err != nil {
		fmt.Printf("sigHeader: %v", sigHeader)
		fmt.Printf("signature: %v", err.Error())
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.async_payment_failed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			fmt.Printf("payload unmarshal: %v", err.Error())
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		orderIDText := strings.TrimSpace(session.Metadata["order_id"])
		if orderIDText == "" {
			fmt.Printf("\nmetadata: %+v", session.Metadata)
			http.Error(w, "Missing order metadata", http.StatusBadRequest)
			return
		}

		orderID, err := strconv.Atoi(orderIDText)
		if err != nil {
			fmt.Printf("\nmetadata: %+v", session.Metadata)
			fmt.Printf("\nmetadata err: %v", err.Error())
			http.Error(w, "Invalid order metadata", http.StatusBadRequest)
			return
		}

		stripePaymentID := session.ID
		if session.PaymentIntent != nil {
			stripePaymentID = session.PaymentIntent.ID
		}

		switch event.Type {
		case "checkout.session.completed", "checkout.session.async_payment_succeeded":
			if err := orders.UpdateOrderPaymentStatus(orderID, "paid", stripePaymentID); err != nil {
				log.Printf("Failed to update payment status: %v", err)
			}
		case "checkout.session.async_payment_failed":
			if err := orders.UpdateOrderPaymentStatus(orderID, "failed", stripePaymentID); err != nil {
				log.Printf("Failed to update payment status: %v", err)
			}
		}

		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

func stripePaymentMethodTypes(method string) ([]*string, error) {
	switch method {
	case "credit_card":
		return stripe.StringSlice([]string{"card"}), nil
	case "boleto":
		return stripe.StringSlice([]string{"boleto"}), nil
	case "pix":
		return stripe.StringSlice([]string{"pix"}), nil
	default:
		return nil, ValidationError{Field: "paymentMethod", Message: "Forma de pagamento inválida"}
	}
}

func stripeLineItems(items []orders.CartItem) ([]*stripe.CheckoutSessionLineItemParams, error) {
	if len(items) == 0 {
		return nil, ValidationError{Field: "cart", Message: "Seu carrinho está vazio"}
	}

	lineItems := make([]*stripe.CheckoutSessionLineItemParams, 0, len(items))
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, ValidationError{Field: "cart", Message: "Quantidade inválida no carrinho"}
		}
		if item.Price <= 0 {
			return nil, ValidationError{Field: "cart", Message: "Preço inválido no carrinho"}
		}

		unitAmount := int64(math.Round(item.Price * 100))
		if unitAmount < 1 {
			return nil, ValidationError{Field: "cart", Message: "Preço inválido no carrinho"}
		}

		lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency:   stripe.String(string(stripe.CurrencyBRL)),
				UnitAmount: stripe.Int64(unitAmount),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name),
				},
			},
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	return lineItems, nil
}
