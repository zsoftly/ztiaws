// Jenkins Pipeline - IAM Keys Authentication Example
//
// This Jenkinsfile demonstrates how to use ztictl in Jenkins with IAM access keys.
// For self-hosted Jenkins on EC2, consider using EC2 instance profiles instead.
//
// Prerequisites:
// 1. Create an IAM user with programmatic access
// 2. Attach necessary permissions to the user (see docs/IAM_PERMISSIONS.md)
// 3. Store credentials in Jenkins:
//    - Credentials ID: aws-access-key-id (Secret text)
//    - Credentials ID: aws-secret-access-key (Secret text)
//    - Optionally: production-instance-id (Secret text)
// 4. Install required plugins:
//    - Pipeline plugin
//    - Credentials plugin
//    - Credentials Binding plugin
//
// Security Note:
// - IAM access keys are long-lived credentials (security risk)
// - Rotate keys regularly (every 90 days minimum)
// - Use EC2 instance profiles if Jenkins runs on AWS
// - Consider using OIDC if available in your Jenkins version

pipeline {
    agent any

    environment {
        AWS_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_NON_INTERACTIVE = 'true'
        ZTICTL_LOG_ENABLED = 'false'
    }

    stages {
        stage('Setup') {
            steps {
                script {
                    echo 'Installing ztictl...'
                    sh '''
                        curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
                        chmod +x ztictl
                        sudo mv ztictl /usr/local/bin/
                        ztictl --version
                    '''

                    echo 'Installing AWS Session Manager plugin...'
                    sh '''
                        curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" \
                          -o "session-manager-plugin.deb"
                        sudo dpkg -i session-manager-plugin.deb
                        session-manager-plugin --version
                    '''
                }
            }
        }

        stage('Configure') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        echo 'Verifying AWS credentials...'
                        sh 'aws sts get-caller-identity'

                        echo 'Initializing ztictl configuration...'
                        sh 'ztictl config init --non-interactive'

                        echo 'Verifying ztictl configuration...'
                        sh 'ztictl config check'
                    }
                }
            }
        }

        stage('List Instances') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        echo 'Listing EC2 instances...'
                        sh 'ztictl ssm list --region ${AWS_DEFAULT_REGION} --table'
                    }
                }
            }
        }

        stage('Deploy to Production') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        echo 'Deploying to production instance...'
                        sh '''
                            ztictl ssm exec ${AWS_DEFAULT_REGION} web-server-prod \
                              "cd /app && git pull && systemctl restart app"
                        '''
                    }
                }
            }
        }

        stage('Deploy to All Web Servers') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        echo 'Deploying to all web servers...'
                        sh '''
                            ztictl ssm exec-tagged ${AWS_DEFAULT_REGION} \
                              --tags Environment=production,Service=web \
                              "/app/deploy.sh"
                        '''
                    }
                }
            }
        }

        stage('Verify Deployment') {
            when {
                branch 'main'
            }
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        echo 'Verifying deployment...'
                        sh '''
                            ztictl ssm exec ${AWS_DEFAULT_REGION} web-server-prod \
                              "systemctl status app && curl -f http://localhost:8080/health"
                        '''
                    }
                }
            }
        }
    }

    post {
        always {
            echo 'Pipeline completed'
        }
        success {
            echo 'Deployment successful'
        }
        failure {
            echo 'Deployment failed'
            // Download logs for debugging
            withCredentials([
                string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY'),
                string(credentialsId: 'production-instance-id', variable: 'INSTANCE_ID')
            ]) {
                script {
                    sh '''
                        mkdir -p artifacts
                        ztictl ssm transfer download \
                          ${INSTANCE_ID} \
                          /var/log/app/error.log \
                          ./artifacts/error.log \
                          --region ${AWS_DEFAULT_REGION} || true
                    '''
                    archiveArtifacts artifacts: 'artifacts/**', allowEmptyArchive: true
                }
            }
        }
    }
}

// Alternative: EC2 Instance Profile Example
// Use this if Jenkins runs on an EC2 instance with an attached IAM role

/*
pipeline {
    agent { label 'ec2' }  // EC2-based Jenkins agent

    environment {
        AWS_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_NON_INTERACTIVE = 'true'
    }

    stages {
        stage('Setup') {
            steps {
                sh '''
                    curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
                    chmod +x ztictl
                    sudo mv ztictl /usr/local/bin/
                '''
            }
        }

        stage('Configure') {
            steps {
                // No credential management needed - uses instance profile
                sh 'aws sts get-caller-identity'
                sh 'ztictl config init --non-interactive'
            }
        }

        stage('Deploy') {
            steps {
                sh '''
                    ztictl ssm exec-tagged ${AWS_DEFAULT_REGION} \
                      --tags Environment=production \
                      "/app/deploy.sh"
                '''
            }
        }
    }
}
*/

// Multi-Region Deployment Example

/*
pipeline {
    agent any

    parameters {
        choice(
            name: 'DEPLOY_REGIONS',
            choices: ['ca-central-1', 'us-east-1', 'eu-west-1', 'all'],
            description: 'Select deployment region(s)'
        )
    }

    environment {
        ZTICTL_NON_INTERACTIVE = 'true'
    }

    stages {
        stage('Setup') {
            steps {
                sh '''
                    curl -L https://github.com/zsoftly/ztiaws/releases/latest/download/ztictl-linux-amd64 -o ztictl
                    chmod +x ztictl
                    sudo mv ztictl /usr/local/bin/
                '''
            }
        }

        stage('Deploy Multi-Region') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    script {
                        def regions = params.DEPLOY_REGIONS == 'all' ?
                            ['ca-central-1', 'us-east-1', 'eu-west-1'] :
                            [params.DEPLOY_REGIONS]

                        regions.each { region ->
                            echo "Deploying to ${region}..."
                            sh """
                                export AWS_DEFAULT_REGION=${region}
                                export ZTICTL_DEFAULT_REGION=${region}
                                ztictl config init --non-interactive
                                ztictl ssm exec-tagged ${region} \
                                  --tags Environment=production \
                                  "/app/deploy.sh"
                            """
                        }
                    }
                }
            }
        }
    }
}
*/

// Test Environment Management Example

/*
pipeline {
    agent any

    environment {
        AWS_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_DEFAULT_REGION = 'ca-central-1'
        ZTICTL_NON_INTERACTIVE = 'true'
    }

    stages {
        stage('Start Test Instances') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    sh '''
                        ztictl ssm start-tagged \
                          --region ${AWS_DEFAULT_REGION} \
                          --tags Environment=test,AutoStart=true
                        echo "Waiting for instances to be ready..."
                        sleep 60
                    '''
                }
            }
        }

        stage('Run Tests') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    sh '''
                        ztictl ssm exec-tagged ${AWS_DEFAULT_REGION} \
                          --tags Environment=test \
                          "cd /app && npm test"
                    '''
                }
            }
        }

        stage('Stop Test Instances') {
            steps {
                withCredentials([
                    string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                    string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
                ]) {
                    sh '''
                        ztictl ssm stop-tagged \
                          --region ${AWS_DEFAULT_REGION} \
                          --tags Environment=test
                    '''
                }
            }
        }
    }

    post {
        always {
            // Always stop test instances, even if tests fail
            withCredentials([
                string(credentialsId: 'aws-access-key-id', variable: 'AWS_ACCESS_KEY_ID'),
                string(credentialsId: 'aws-secret-access-key', variable: 'AWS_SECRET_ACCESS_KEY')
            ]) {
                sh '''
                    ztictl ssm stop-tagged \
                      --region ${AWS_DEFAULT_REGION} \
                      --tags Environment=test || true
                '''
            }
        }
    }
}
*/
