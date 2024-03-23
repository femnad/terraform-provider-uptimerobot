resource "uptimerobot_alert_contact" "example" {
  friendly_name = "email-contact"
  type          = "e-mail"
  value         = "alert@example.com"
}
