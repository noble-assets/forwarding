alias forwardingd=./simapp/build/simd

for arg in "$@"
do
    case $arg in
        -r|--reset)
        rm -rf .forwarding
        shift
        ;;
    esac
done

if ! [ -f .forwarding/data/priv_validator_state.json ]; then
  forwardingd init validator --chain-id "forwarding-1" --home .forwarding &> /dev/null

  forwardingd keys add validator --home .forwarding --keyring-backend test &> /dev/null
  forwardingd genesis add-genesis-account validator 1000000ustake --home .forwarding --keyring-backend test
  forwardingd keys add authority --recover --home .forwarding --keyring-backend test <<< "exchange flash debris claw calm shine laundry february cousin glad name miss jar neglect reflect blanket orbit clever educate rent inject lounge pupil plastic" &> /dev/null
  forwardingd genesis add-genesis-account authority 10000000uusdc --home .forwarding --keyring-backend test
  forwardingd keys add user --home .forwarding --keyring-backend test &> /dev/null
  forwardingd genesis add-genesis-account user 10000000uusdc --home .forwarding --keyring-backend test

  TEMP=.forwarding/genesis.json
  touch $TEMP && jq '.app_state.staking.params.bond_denom = "ustake"' .forwarding/config/genesis.json > $TEMP && mv $TEMP .forwarding/config/genesis.json

  forwardingd genesis gentx validator 1000000ustake --chain-id "forwarding-1" --home .forwarding --keyring-backend test &> /dev/null
  forwardingd genesis collect-gentxs --home .forwarding &> /dev/null

  sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' .forwarding/config/config.toml
fi

forwardingd start --home .forwarding
