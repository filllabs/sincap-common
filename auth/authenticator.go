package auth

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"gitlab.com/sincap/sincap-common/auth/claims"
	"gitlab.com/sincap/sincap-common/net"
)

// CheckTokenOwnership validates the request information with the info inside of the claims

var bypassIPMap map[string]bool

func CheckTokenOwnership(ctx *fiber.Ctx, config Config, c *claims.DecryptedClaims) error {
	currentIP := net.ReadUserIP(ctx)

	// Convert OwnershipBypassIPss to map for faster lookup
	// Convert and check ops only will happen any OwnershipBypassIPss given
	if config.OwnershipBypassIPs != nil && len(config.OwnershipBypassIPs) > 0 {
		// Create map of OwnershipBypassIPss if not created yet
		if bypassIPMap == nil {
			// fill from slice
			bypassIPMap = make(map[string]bool, len(config.OwnershipBypassIPs))
			for _, ip := range config.OwnershipBypassIPs {
				bypassIPMap[ip] = true
			}
		}
		// Check if request IP is in OwnershipBypassIPs
		if _, ok := bypassIPMap[net.ReadUserIP(ctx)]; ok {
			return nil
		}
	}

	// if DeviceIDCheck is true, check if device ID is the same with the tokens device id
	if config.DeviceIDCheck {
		if cookie := ctx.Cookies("X-Device"); cookie != c.Extra["X-Device"] {
			return fmt.Errorf("token error, deviceid info is not equal. User deviceid : %v  Claims deviceid : %v", ctx.Cookies("X-Device"), c.Extra["X-Device"])
		}
	}
	if config.IPCheck {
		if currentIP != c.UserIP {
			return fmt.Errorf("token error, ip is not equal. User ip : %v  Claims ip : %v", currentIP, c.UserIP)
		}
	}

	return nil
}

// RenewTokenIfNeeded renews the token if it is near to the expiration time
func RenewTokenIfNeeded(dclaims *claims.DecryptedClaims, ctx *fiber.Ctx, config Config) error {
	// if any value given start renewing.
	if config.RefreshBefore > 0 {
		// if close to expiration give new token
		untilExpire := dclaims.ExpiresAt - time.Now().UTC().Unix()
		if untilExpire < config.RefreshBefore {

			// update expire time
			dclaims.ExpiresAt = config.Timeout + time.Now().UTC().Unix()
			eclaims, err := dclaims.Encrypt(config.Secret)
			if err != nil {
				return ctx.SendStatus(fiber.StatusInternalServerError)
			}
			// create new token
			token := jwt.New(jwt.SigningMethodHS256)
			token.Claims = eclaims
			tokenString, err := token.SignedString([]byte(config.Secret))
			if err != nil {
				return ctx.SendStatus(fiber.StatusInternalServerError)
			}
			// set new token as cookie
			ctx.Cookie(&fiber.Cookie{
				Name:     "jwt",
				Value:    tokenString,
				Path:     "/",
				MaxAge:   int(config.Timeout),
				HTTPOnly: config.HTTPOnly,
				Secure:   config.Secure,
				SameSite: fiber.CookieSameSiteLaxMode,
			})
		}
	}
	return nil
}

func InvalidateCookies(ctx *fiber.Ctx, config Config) {
	jwt := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		MaxAge:   int(config.Timeout),
		HTTPOnly: config.HTTPOnly,
		Secure:   config.Secure,
		SameSite: fiber.CookieSameSiteLaxMode,
	}
	ctx.Cookie(&jwt)

	loggedin := fiber.Cookie{
		Name:     "loggedin",
		Value:    "false",
		Path:     "/",
		MaxAge:   int(config.Timeout),
		SameSite: fiber.CookieSameSiteLaxMode,
	}
	ctx.Cookie(&loggedin)
}
