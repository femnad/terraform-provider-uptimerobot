# Lookup alert contact with the given friendly name.
data "uptimerobot_alert_contact" "exa1" {
  friendly_name = "exa1"
}

# Lookup alert contact with the given type and email (must match both).
data "uptimerobot_alert_contact" "exa2" {
  type  = "e-mail"
  value = "exa2@example.com"
}
