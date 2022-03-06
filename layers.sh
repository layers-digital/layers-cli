# DEPRECATED
# use the golang files instead


LAYERS_DIR=('layers-comunicados' 'layers-atendimentos' 'tendaedu-web' 'tendaedu-backend')

layers-cli() {
  echo "hello"
}

layers_dir() {
  SRC=$PWD
  BASE=${SRC##*/}

  echo $BASE
}

layers_help() {
  echo "HELP"
}

layers_node() {
  NODE_VERSION="$(node --version)"

  # check in nvmrc
  if [[ -e .nvmrc ]]; then
    DIR_VERSION="$(cat .nvmrc)"
    SHORT="${DIR_VERSION::3}"
    MAJOR_NODEVER="${NODE_VERSION::3}"
    echo ${SHORT}
    if [[ ${MAJOR_NODEVER} == ${SHORT} ]]; then
      echo "okay"
    else
      echo "wrong ver"
    fi
  fi
}


main() {
  COMMANDS=('dir' 'node' 'help')
  echo
  if [[ -n "$1" ]]; then
    if [[ " ${COMMANDS[@]} " =~ " ${1} " ]]; then
      layers_$1
    else 
      echo "'$1' command not found"
    fi
  else
    echo "show help"
  fi
}

main $*