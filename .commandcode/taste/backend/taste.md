# Backend Preferences

- Prefers Keycloak (OIDC) as the production authentication method over local JWT, while keeping both auth modes working. Confidence: 0.8
- Prefers WebSocket-based real-time communication end-to-end over HTTP polling (e.g., payment/status updates). Confidence: 0.8
- Config and secrets must come from environment variables, never hardcoded (e.g., static QRIS strings, payment gateway keys). Confidence: 0.8
- In the payment domain model, a void is an explicit, cashier-triggered action and must not be conflated with a failed transaction — failed payments need to stay distinguishable from voided orders (e.g., don't treat auto-void-on-failure as equivalent to an explicit void). Confidence: 0.8
