from drgn.env import config_from_env

if __name__ == "__main__":
    print("Running entrypoint of orderbook")
    config_from_env()

    from main import start_multiple_traders

    start_multiple_traders(10)