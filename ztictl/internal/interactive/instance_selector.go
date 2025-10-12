package interactive

import (
	"fmt"

	"ztictl/pkg/colors"

	"github.com/fatih/color"
)

// Instance represents an EC2 instance with SSM information
type Instance struct {
	InstanceID       string
	Name             string
	State            string
	Platform         string
	PrivateIPAddress string
	PublicIPAddress  string
	SSMStatus        string
	SSMAgentVersion  string
	LastPingDateTime string
	Tags             map[string]string
}

// InstanceSelector is an interface for selecting an instance.
type InstanceSelector interface {
	SelectInstance(instances []Instance) (*Instance, error)
}

// FuzzyInstanceSelector is a fuzzy finder for selecting an instance.
type FuzzyInstanceSelector struct{}

// SelectInstance uses a fuzzy finder to select an instance from a list.
func (s *FuzzyInstanceSelector) SelectInstance(instances []Instance) (*Instance, error) {
	return SelectInstance(instances, "Select EC2 Instance")
}

// SelectInstance provides a common interface for instance selection
func SelectInstance(instances []Instance, title string) (*Instance, error) {
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances available")
	}

	idx, err := FuzzyFind(instances,
		func(i int) string {
			instance := instances[i]
			name := instance.Name
			if name == "" {
				name = "N/A"
			}
			return fmt.Sprintf("%s (%s)", name, instance.InstanceID)
		},
		fmt.Sprintf("%s (%d available)", title, len(instances)),
		func(i, w, h int) string {
			if i < 0 || i >= len(instances) {
				return ""
			}
			instance := instances[i]
			name := instance.Name
			if name == "" {
				name = "N/A"
			}

			var ssmStatus string
			switch instance.SSMStatus {
			case "Online":
				ssmStatus = colors.ColorSuccess("✓ Online")
			case "ConnectionLost":
				ssmStatus = colors.ColorWarning("⚠ Lost")
			case "No Agent":
				ssmStatus = colors.ColorError("✗ No Agent")
			default:
				if instance.SSMStatus == "" {
					ssmStatus = colors.ColorError("✗ No Agent")
				} else {
					ssmStatus = colors.ColorWarning("? %s", instance.SSMStatus)
				}
			}

			publicIP := instance.PublicIPAddress
			if publicIP == "" {
				publicIP = "N/A"
			}

			return fmt.Sprintf("Name:         %s\n"+
				"Instance ID:  %s\n"+
				"State:        %s\n"+
				"Platform:     %s\n"+
				"Private IP:   %s\n"+
				"Public IP:    %s\n"+
				"SSM Status:   %s",
				name, instance.InstanceID, instance.State, instance.Platform, instance.PrivateIPAddress, publicIP, ssmStatus)
		},
	)

	if err != nil {
		if err.Error() == "abort" {
			color.New(color.FgRed).Println("❌ Instance selection cancelled")
			return nil, fmt.Errorf("instance selection cancelled")
		}
		return nil, fmt.Errorf("instance selection failed: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Selected: %s (%s)\n", instances[idx].Name, instances[idx].InstanceID)

	return &instances[idx], nil
}
