resource "mailjet_sender" "sender_example" {
  email = "*@mailjet.example.com"
  name = "My mailjet sender example"
  is_default_sender = false
  email_type = "unknown"
}

data "mailjet_dns" "example" {
  dns_id = resource.mailjet_sender.sender_example.dns_id
}
