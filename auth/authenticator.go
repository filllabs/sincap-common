package auth

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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

	if currentIP != c.UserIP {
		return fmt.Errorf("token error, ip is not equal. User ip : %v  Claims ip : %v", currentIP, c.UserIP)
	}
	return nil
}
