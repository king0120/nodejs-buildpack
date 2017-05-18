create_default_env() {
  export NPM_CONFIG_PRODUCTION=${NPM_CONFIG_PRODUCTION:-true}
  export NPM_CONFIG_LOGLEVEL=${NPM_CONFIG_LOGLEVEL:-error}
  export NODE_MODULES_CACHE=${NODE_MODULES_CACHE:-true}
  export NODE_ENV=${NODE_ENV:-production}
  export NODE_VERBOSE=${NODE_VERBOSE:-false}
  export LD_LIBRARY_PATH="$DEPS_DIR/$DEPS_IDX/node/instantclientbasic"
  export OCI_INC_DIR="$DEPS_DIR/$DEPS_IDX/node/instantclientbasic/sdk/include"
  export OCI_LIB_DIR="$DEPS_DIR/$DEPS_IDX/node/instantclientbasic"
}

list_node_config() {
  echo ""
  printenv | grep ^NPM_CONFIG_ || true
  printenv | grep ^YARN_ || true
  printenv | grep ^NODE_ || true

  if [ "$NPM_CONFIG_PRODUCTION" = "true" ] && [ "$NODE_ENV" != "production" ]; then
    echo ""
    echo "npm scripts will see NODE_ENV=production (not '${NODE_ENV}')"
    echo "https://docs.npmjs.com/misc/config#production"
  fi
}

