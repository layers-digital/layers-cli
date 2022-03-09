case "$1" in
  -b|--build)
    go build
    echo "built layers_cli"
    ;;
esac

if test -f ~/.bashrc; then
    RC_FILE=~/.bashrc
fi

if test -f ~/.zshrc; then
    RC_FILE=~/.zshrc
fi

if grep -q "alias layers=" "$RC_FILE"; then
  echo "Already installed at $RC_FILE"
  exit
fi

echo "" >> $RC_FILE
echo "# Layers CLI" >> $RC_FILE
echo "alias layers=\"$PWD/layers_cli\"" >> $RC_FILE

echo "Restart this terminal and run 'layers'"