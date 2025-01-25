variable "disable_bgp_route_propagation" {
  description = "Controls propagation of routes learned by BGP on that route table."
  type        = bool
  default     = true
  nullable    = false
}
