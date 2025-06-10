module fem-broker

go 1.21

require github.com/fep-fem/protocol v0.0.0

require github.com/golang-jwt/jwt/v5 v5.2.0 // indirect

replace github.com/fep-fem/protocol => ../protocol/go
