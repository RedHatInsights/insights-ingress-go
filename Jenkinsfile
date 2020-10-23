
pipeline {
  agent {
    node {
      label 'golang112'
    }
  }
  stages {
    stage('Go Test') {
      steps {
        sh 'export GO111MODULE="on"'
        sh 'export GOPATH=/var/gopath'
        sh 'go version'
        sh 'export ACG_CONFIG=$(pwd)/cdappconfig.json'
        sh 'ACG_CONFIG=$ACG_CONFIG go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...'
      }
    }    
  }
}

