from drgn.env import config_from_env

if __name__ == "__main__":
    print("Running entrypoint of traderpool")
    config_from_env()

    from main import start

    start()