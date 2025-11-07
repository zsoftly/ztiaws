<#
.SYNOPSIS
    Manages AWS Systems Manager (SSM) sessions across different AWS regions.

.DESCRIPTION
    This PowerShell function simplifies AWS Systems Manager (SSM) session management.
    It can list available EC2 instances that are manageable through SSM in a specified region
    and start SSM sessions to connect to specific instances.

.PARAMETER regionAlias
    The region alias to connect to. Supported values:
    - cac1: Canada Central (ca-central-1)
    - use2: US East Ohio (us-east-2)
    - euc1: Europe Frankfurt (eu-central-1)

.PARAMETER instanceId
    The ID of the EC2 instance to connect to (e.g., i-1234567890abcd)

.EXAMPLE
    ssm cac1
    Lists all SSM-enabled instances in Canada Central region

.EXAMPLE
    ssm use2 i-1234567890abcd
    Connects to the specified instance in US East (Ohio) region

.NOTES
    Requires:
    - AWS CLI to be installed and configured
    - Appropriate AWS permissions to access SSM and EC2 services
#>
function ssm {
    param (
        [string]$regionAlias,
        [string]$instanceId
    )
    # Function to show usage
    function Show-Usage {
        Write-Host "Usage: ssm [region] [instance-id]"
        Write-Host "Supported regions:"
        Write-Host "  cac1  - Canada (Central)"
        Write-Host "  use2  - US East (Ohio)"
        Write-Host "  euc1  - Europe (Frankfurt)"
        Write-Host "Examples:"
        Write-Host "  ssm cac1         # List instances in Canada Central"
        Write-Host "  ssm use2 i-1234  # Connect to instance in US East (Ohio)"
    }

    # Convert region alias to full region name
    function Get-AWSRegion {
        param ([string]$alias)
        
        switch ($alias) {
            "cac1" { "ca-central-1" }
            "use2" { "us-east-2" }
            "euc1" { "eu-central-1" }
            default { 
                if (!$alias) { "ca-central-1" }  # Default to Canada Central
                else { "invalid" }
            }
        }
    }

    # If no parameters, show usage
    if (!$regionAlias -and !$instanceId) {
        Show-Usage
        return
    }

    # Get region from alias or use default
    $region = Get-AWSRegion $regionAlias
    if ($region -eq "invalid") {
        Write-Host "Error: Invalid region. Use cac1, use2, or euc1" -ForegroundColor Red
        Show-Usage
        return
    }

    if (!$instanceId) {
        # List instances in specified region
        Write-Host "Available instances in ${region}:" -ForegroundColor Green
        
        # First get instance info from SSM
        $ssmInstances = aws ssm describe-instance-information `
            --region ${region} `
            --output json | ConvertFrom-Json

        # Get the instance IDs
        $instanceIds = $ssmInstances.InstanceInformationList.InstanceId

        # Get tags for these instances
        $instanceDetails = aws ec2 describe-instances `
            --region ${region} `
            --instance-ids $instanceIds `
            --query 'Reservations[].Instances[].[Tags[?Key==`Name`].Value|[0],InstanceId]' `
            --output json | ConvertFrom-Json

        # Create a hashtable for quick instance name lookup
        $nameMap = @{}
        foreach ($instance in $instanceDetails) {
            $nameMap[$instance[1]] = if ($instance[0]) { $instance[0] } else { "Unnamed" }
        }

        # Format and display the results
        $ssmInstances.InstanceInformationList | ForEach-Object {
            $name = $nameMap[$_.InstanceId]
            $platform = if ($_.PlatformName -like "*Microsoft*") { "Windows" } else { "Linux" }
            [PSCustomObject]@{
                Name = $name
                InstanceId = $_.InstanceId
                Status = $_.PingStatus
                Platform = $platform
            }
        } | Format-Table -AutoSize

        Write-Host "`nUsage: ssm ${regionAlias} <instance-id>" -ForegroundColor Yellow
    }
    else {
        # Connect to instance in specified region using default cac1 if no region specified
        if (!$regionAlias) {
            $region = "ca-central-1"  # Default to cac1 only when connecting
        }
        aws ssm start-session `
            --region ${region} `
            --target $instanceId
    }
}
