version=$1
sample_yaml=$2

exist_crd=$(kubectl get crd | grep mysql.radondb.com)
exist_operator_name=$(helm list -A | grep mysql-operator | awk '{print $1}')
exist_operator_ns=$(helm list -A | grep mysql-operator | awk '{print $2}')

if [ -n "$exist_crd" ]; then
    sure_delete 'crds'
    if [ $? -eq 1 ]; then
        delete_crds
        check_crd=$(kubectl get crd | grep mysql.radondb.com)
        if [ ! -n "$check_crd" ]; then
            echo "delete crds success"
            install_radondb_mysql $version
        else
            echo "failed to delete all crds"
        fi
    fi
else
    install_radondb_mysql $version
fi

install_radondb_mysql(){
    echo "installing ..."
    add_repo
    install_operator $1
    install_mysqlcluster $1
    echo "Welcome to RadonDB MySQL($1)"
    echo "More doc: https://github.com/radondb/radondb-mysql-kubernetes#features"
}

install_mysqlcluster(){
    if [ -n "$sample_yaml" ]; then
        kubectl apply -f $sample_yaml
    else
cat <<EOF | kubectl apply -f-
apiVersion: mysql.radondb.com/v1alpha1
kind: MysqlCluster
metadata:
  name: sample
spec:
  replicas: 3
EOF
    fi
}

create_operator_namespace(){
    kubectl create ns radondb-mysql
}

install_operator(){
    if [ -n "$exist_operator_name" ]; then
        sure_delete 'operator'
        if [ $? -eq 1 ]; then 
            helm delete $exist_operator_name -n $exist_operator_ns
        fi
    fi
    case $1 in
    "v2.1.0")
        helm install demo radondb/mysql-operator --version 0.1.0 -n radondb-mysql
        ;;
    "v2.1.1")
        helm install demo radondb/mysql-operator --version 0.1.1 -n radondb-mysql
        ;;
    *)
    echo "invalid version"
    esac
    sleep 5
}

add_repo(){
    helm repo add radondb https://radondb.github.io/radondb-mysql-kubernetes/ && helm repo update
}

delete_crds(){
    kubectl delete crd mysqlclusters.mysql.radondb.com
    kubectl delete crd backups.mysql.radondb.com
    kubectl delete crd mysqlusers.mysql.radondb.com
    sleep 5
}

sure_delete(){
    read -r -p "delete exist $1? [Y/n] " input
    case $input in
        [yY][eE][sS]|[yY])
            return 1
            ;;

        [nN][oO]|[nN])
            return 0
            ;;
        *)
        echo "Invalid input..."
        return 0
        ;;
    esac
}


