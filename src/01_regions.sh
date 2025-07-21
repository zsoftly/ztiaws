#!/usr/bin/env bash

# AWS Regions mapping (full coverage)
# Format: "shortcode") echo "aws-region-name" ;;  # Location/City

get_region() {
    case "$1" in
        # Canada
        "cac1") echo "ca-central-1" ;;      # Canada (Central) - Montreal
        "caw1") echo "ca-west-1" ;;         # Canada (West) - Calgary

        # United States
        "use1") echo "us-east-1" ;;          # US East (N. Virginia)
        "use2") echo "us-east-2" ;;          # US East (Ohio)
        "usw1") echo "us-west-1" ;;          # US West (N. California)
        "usw2") echo "us-west-2" ;;          # US West (Oregon)
        "usgov1") echo "us-gov-west-1" ;;    # AWS GovCloud (US-West)
        "usgov2") echo "us-gov-east-1" ;;    # AWS GovCloud (US-East)

        # Europe
        "euw1") echo "eu-west-1" ;;          # EU (Ireland)
        "euw2") echo "eu-west-2" ;;          # EU (London)
        "euw3") echo "eu-west-3" ;;          # EU (Paris)
        "euc1") echo "eu-central-1" ;;       # EU (Frankfurt)
        "eun1") echo "eu-north-1" ;;         # EU (Stockholm)
        "eum1") echo "eu-south-1" ;;         # EU (Milan)
        "eub1") echo "eu-west-4" ;;          # EU (Birmingham)

        # Asia Pacific
        "apse1") echo "ap-southeast-1" ;;    # Asia Pacific (Singapore)
        "apse2") echo "ap-southeast-2" ;;    # Asia Pacific (Sydney)
        "apse3") echo "ap-southeast-3" ;;    # Asia Pacific (Jakarta)
        "apne1") echo "ap-northeast-1" ;;    # Asia Pacific (Tokyo)
        "apne2") echo "ap-northeast-2" ;;    # Asia Pacific (Seoul)
        "apne3") echo "ap-northeast-3" ;;    # Asia Pacific (Osaka)
        "aps1") echo "ap-south-1" ;;         # Asia Pacific (Mumbai)
        "apme1") echo "me-south-1" ;;        # Middle East (Bahrain)
        "apse4") echo "ap-east-1" ;;         # Asia Pacific (Hong Kong)
        "apne4") echo "ap-northeast-4" ;;    # Asia Pacific (Hyderabad)

        # South America
        "sae1") echo "sa-east-1" ;;          # South America (São Paulo)

        # Africa
        "afsa1") echo "af-south-1" ;;        # Africa (Cape Town)

        # China (requires special access)
        "cn-n1") echo "cn-north-1" ;;        # China (Beijing)
        "cn-n2") echo "cn-northwest-1" ;;    # China (Ningxia)

        # AWS Local Zones
        "usw4") echo "us-west-2-lax-1" ;;    # US West Local Zone (Los Angeles)
        "use4") echo "us-east-1-bos-1" ;;    # US East Local Zone (Boston)
        "use5") echo "us-east-1-dca-1" ;;    # US East Local Zone (Washington DC)
        "use6") echo "us-east-1-phi-1" ;;    # US East Local Zone (Philadelphia)

        # Default case
        *) echo "invalid" ;;
    esac
}

get_region_description() {
    case "$1" in
        "cac1") echo "Canada Central (Montreal)" ;;
        "caw1") echo "Canada West (Calgary)" ;;
        "use1") echo "US East (N. Virginia)" ;;
        "use2") echo "US East (Ohio)" ;;
        "usw1") echo "US West (N. California)" ;;
        "usw2") echo "US West (Oregon)" ;;
        "usgov1") echo "US GovCloud (West)" ;;
        "usgov2") echo "US GovCloud (East)" ;;
        "euw1") echo "EU West (Ireland)" ;;
        "euw2") echo "EU West (London)" ;;
        "euw3") echo "EU West (Paris)" ;;
        "euc1") echo "EU Central (Frankfurt)" ;;
        "eun1") echo "EU North (Stockholm)" ;;
        "eum1") echo "EU South (Milan)" ;;
        "eub1") echo "EU West (Birmingham)" ;;
        "apse1") echo "Asia Pacific (Singapore)" ;;
        "apse2") echo "Asia Pacific (Sydney)" ;;
        "apse3") echo "Asia Pacific (Jakarta)" ;;
        "apne1") echo "Asia Pacific (Tokyo)" ;;
        "apne2") echo "Asia Pacific (Seoul)" ;;
        "apne3") echo "Asia Pacific (Osaka)" ;;
        "aps1") echo "Asia Pacific (Mumbai)" ;;
        "apme1") echo "Middle East (Bahrain)" ;;
        "apse4") echo "Asia Pacific (Hong Kong)" ;;
        "apne4") echo "Asia Pacific (Hyderabad)" ;;
        "sae1") echo "South America (São Paulo)" ;;
        "afsa1") echo "Africa (Cape Town)" ;;
        "cn-n1") echo "China (Beijing)" ;;
        "cn-n2") echo "China (Ningxia)" ;;
        "usw4") echo "US West Local Zone (Los Angeles)" ;;
        "use4") echo "US East Local Zone (Boston)" ;;
        "use5") echo "US East Local Zone (Washington DC)" ;;
        "use6") echo "US East Local Zone (Philadelphia)" ;;
        *) echo "Unknown Region" ;;
    esac
}

# Optional: list all supported region shortcodes
list_region_shortcodes() {
    cat <<EOF
cac1  caw1  use1  use2  usw1  usw2  usgov1  usgov2
euw1  euw2  euw3  euc1  eun1  eum1  eub1
apse1 apse2 apse3 apne1 apne2 apne3 aps1 apme1 apse4 apne4
sae1  afsa1
cn-n1 cn-n2
usw4  use4  use5  use6
EOF
}
