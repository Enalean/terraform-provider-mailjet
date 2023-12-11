resource "mailjet_sender" "sender_example_specific" {
  email = "my_email@mailjet.example.com"
  name = "My mailjet sender example for a specific email"
  is_default_sender = false
  email_type = "unknown"
}

resource "mailjet_sender" "sender_example_whole_domain" {
  email = "*@mailjet.example.com"
  name = "My mailjet sender example for a whole domain"
  is_default_sender = false
  email_type = "unknown"
}
