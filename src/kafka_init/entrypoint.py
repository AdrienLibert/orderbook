import argparse

from drgn.env import config_from_env

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--init-topics", dest="init_topics", action="store_true")
    parser.add_argument("--init-orderbook", dest="init_orderbook", action="store_true")
    args = parser.parse_args()

    script = "init_topics" if args.init_topics else "init_orderbook"

    print(f"Running entrypoint of kafka init: {args}")
    config_from_env()

    from main import main

    main(script)
