#!/usr/bin/env bash

# AWS Regions mapping
# Format: "shortcode") echo "aws-region-name" ;;  # Location/City

get_region() {
    case "$1" in
        # Canada
        "cac1") echo "ca-central-1" ;;  # Montreal
        "caw1") echo "ca-west-1" ;;     # Calgary

        # United States
        "use1") echo "us-east-1" ;;     # N. Virginia
        "use2") echo "us-east-2" ;;     # Ohio
        "usw1") echo "us-west-1" ;;     # N. California
        "usw2") echo "us-west-2" ;;     # Oregon

        # Europe
        "euw1") echo "eu-west-1" ;;     # Ireland
        "euw2") echo "eu-west-2" ;;     # London
        "euw3") echo "eu-west-3" ;;     # Paris
        "euc1") echo "eu-central-1" ;;  # Frankfurt

        # Asia Pacific
        "aps1") echo "ap-south-1" ;;    # Mumbai
        "apse1") echo "ap-southeast-1" ;; # Singapore
        "apse2") echo "ap-southeast-2" ;; # Sydney
        "apne1") echo "ap-northeast-1" ;; # Tokyo
        "apne2") echo "ap-northeast-2" ;; # Seoul
        "apne3") echo "ap-northeast-3" ;; # Osaka

        # South America
        "sae1") echo "sa-east-1" ;;     # São Paulo

        # Default
        *) echo "invalid" ;;
    esac
}

# Get region description
get_region_description() {
    case "$1" in
        "cac1") echo "Canada Central (Montreal)" ;;
        "caw1") echo "Canada West (Calgary)" ;;
        "use1") echo "US East (N. Virginia)" ;;
        # Add more descriptions as needed
        *) echo "Unknown Region" ;;
    esac
}