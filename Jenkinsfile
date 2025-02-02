pipeline {

    agent any

    environment {
        IMAGE_NAME = 'quay.io/csye7125_webapp/webapp-container-registry/kube-operator'
    }

    triggers {
        GenericTrigger(
             genericVariables: [
                 [key: 'LATEST_RELEASE_TAG', value: '$.release.tag_name']
             ],
             token: 'kube-operator-token'
        )
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm: [
                    $class: 'GitSCM',
                    branches: [[name: '*/master']],
                    doGenerateSubmoduleConfigurations: false,
                    extensions: [[$class: 'CloneOption', noTags: false]],
                    submoduleCfg: [],
                    userRemoteConfigs: [[credentialsId: 'GITHUB_TOKEN',
                    url: 'https://github.com/cyse7125-fall2023-group01/kube-operator.git']]
                ]
            }
        }
        stage('Print Tag Details') {
            steps {
                script {
                    env.LATEST_RELEASE_TAG = sh(returnStdout: true, script: 'git describe --tags --abbrev=0').trim()
                    echo "Latest Release Tag: ${env.LATEST_RELEASE_TAG}"
                }
            }
        }
        stage('Docker Version') {
            steps {
                script {
                    sh 'docker version'
                }
            }
        }
        stage('Build Docker Image') {
            steps {
                script {
                    sh "docker build --no-cache -t ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG} ."
                    sh "docker image tag ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG} ${env.IMAGE_NAME}:latest"
                }
            }
        }
        stage('List Docker Images') {
            steps {
                script {
                    sh 'docker image ls'
                }
            }
        }
        stage('Quay Login') {
            steps {
                script {
                    withCredentials([
                    string(credentialsId: 'quayLogin', variable: 'quayLogin'),
                    string(credentialsId: 'quayEncryptedPwd', variable: 'quayEncryptedPwd')]) {
                        sh 'docker login -u=${quayLogin} -p=${quayEncryptedPwd} quay.io'
                    }
                }
            }
        }
        stage('Push Image To Quay') {
            steps {
                script {
                   withCredentials([string(credentialsId: 'quayEncryptedPwd', variable: 'quayEncryptedPwd')]) {
                        sh "docker push ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG}"
                        sh "docker push ${env.IMAGE_NAME}:latest"
                        sh "docker image rmi ${env.IMAGE_NAME}:${env.LATEST_RELEASE_TAG}"
                    }
                }
            }
        }
        
        stage('connect to cluster') {
            steps {
                script {
                    sh 'gcloud auth activate-service-account --key-file ~/gcp_sa_key.json'
                    sh 'gcloud container clusters get-credentials primary --region us-east1 --project root-mapper-401202'
                    sh 'kubectl get pods'
                    sh 'make deploy IMG=quay.io/csye7125_webapp/webapp-container-registry/kube-operator:latest'
                }
            }
        }
        stage('make deploy') {
            steps {
                script {
                    withCredentials([
                    string(credentialsId: 'quayLogin', variable: 'quayLogin'),
                    string(credentialsId: 'quayEncryptedPwd', variable: 'quayEncryptedPwd')]) {
                        sh 'docker login -u=${quayLogin} -p=${quayEncryptedPwd} quay.io'
                        sh 'make undeploy'
                        sh 'make deploy IMG=quay.io/csye7125_webapp/webapp-container-registry/kube-operator'
                    }
                }
            }
        }

    }
}
