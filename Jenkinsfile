node {

    stage('sjekk ut kode') {
        checkout scm
    }

    stage('kompiler og bygg distribusjon'){
        def byggimage = docker.image("priv-docker.systemfabrikken.org:5000/base/golang-jenkins-bygg:go-1.7.3'")
        byggimage.inside("--name='$env.BUILD_TAG'") {
            sh 'go get -d -v'
            sh 'go build -v -o build/prometheus-postgres-exporter'
        }
        archive 'build/prometheus-postgres-exporter'
    }

}