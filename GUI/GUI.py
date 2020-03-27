from sys import exit
from urllib.request import urlopen, URLError
from json import loads, dumps, load
import tkinter as tk
from time import time, sleep
from base64 import b64decode


class GUI(tk.Tk):
    
    WIDTH = 10
    HEIGHT = 10

    REJECTED_MOVES = {
        "r": 37,
        "d": 38,
        "l": 39,
        "u": 40,
    }
    # Stop the snake from going backwards on itself

    def __init__(self, conf):
        super().__init__()
        # Init the tkinter inherited class

        self.token = ""

        self.title("Snake")

        self.direction = "r"

        self.conf = conf

        self.url = conf.get("server", "http://localhost:8081")

        self.image = None

        self.lock = False

        self.label = tk.Label(
            self,
            image=self.init_grid_img(),
            anchor="center"
        )

        # self.label.bind("<Key>", self.key_handler)
        self.bind_all("<Key>", self.key_handler)
        # bind_all over bind as it fixes issues on having to 
        # focus on a specific widget
        self.label.pack()

        self.create_game()

        sleep(3)

        self.clock(self.clock_event)

        # Start the game clock

    def key_handler(self, evt):
        evt.widget.focus_set()

        if evt.keycode not in self.REJECTED_MOVES.values():
            return

        # If not an arrow key, ignore

        if self.REJECTED_MOVES.get(self.direction, 0) == evt.keycode:
            return
        
        # If the move would cause the snake to go directly back on
        # itself, ignore

        if evt.keycode == 37:
            self.direction = "l"
        elif evt.keycode == 38:
            self.direction = "u"
        elif evt.keycode == 39:
            self.direction = "r"
        elif evt.keycode == 40:
            self.direction = "d"
        
    def create_game(self):
        # Generate token and create game on server
        try:
            res = loads(
                urlopen(
                    self.url + "/start",
                    data=dumps(
                        {
                            "Width": self.WIDTH,
                            "Height": self.HEIGHT
                        }
                    ).encode()
                    # Convert to bytes
                ).read().decode()
                # Convert back to string
            )
            # Parse from JSON structure

            if "Error" in res.keys():
                exit(res["Error"])

            self.token = res["Token"]

        except URLError:
            exit("Unable to connect to server")

    def clock_event(self):
        try:
            res = loads(
                urlopen(
                    self.url + "/move",
                    data=dumps(
                        {
                            "Token": self.token,
                            "Direction": self.direction
                        }
                    ).encode()
                ).read().decode()
            )

            if "Error" in res.keys():
                print("[WARN] Error", res["Error"])
                return

            if not res["Alive"]:
                return self.emit_dead(res)
            
            if res["Won"]:
                return self.emit_won(res)

            self.update_board(res)

        except URLError:
            exit("Unable to connect to server")

    def clock(self, func):
        start_time = time() * 1000
        interval_time = 0.5 * 1000

        # Work in milliseconds to allow use of modulus in sub-second
        # scenarios

        while True:
            self.update()
            # Update tkinter class to keep listening for events
            if self.lock:
                break
                # Ignore further inputs
                
            query_time = start_time + (
                    interval_time - ((time() * 1000) - start_time)
                    % interval_time
            )
            # Keep line length <= 72
            
            if time() * 1000 >= query_time:
                # Tkinter doesn't tend to play nice with threads
                start_time = time() * 1000
                func()
            # Runs once every half a second

        self.mainloop()
        # Keep window responsive

    def init_grid_img(self):
        # Creates an image using the server /image method
        default_board = {
            "BoardPositions": [
                [
                    " " for _ in range(self.WIDTH)
                ] for _ in range(self.HEIGHT)
            ],
            "HeadColour": self.conf.get("HeadColour", [
                0,
                0,
                255,
            ]),
            "BodyColour": self.conf.get("BodyColour", [
                50,
                130,
                170,
            ]),
            "AppleColour": self.conf.get("AppleColour", [
                255,
                0,
                0,
            ]),
            "BackgroundColour": self.conf.get("BackgroundColour", [
                60,
                170,
                50
            ]),
            "BlockHeight": self.conf.get("BlockHeight", 50),
            "BlockWidth": self.conf.get("BlockWidth", 50),
            "BorderColour": self.conf.get("BorderColour", [
                255,
                255,
                255
            ])
        }
        
        try:
            res = loads(
                urlopen(
                    self.url + "/image",
                    data=dumps(default_board).encode(),
                ).read().decode()
            )
            # Makes request, returned in JSON format and parsed

            if "Error" in res.keys():
                exit(res.keys.get("Error"))

            self.image = tk.PhotoImage(data=b64decode(res["Image"]))

            return self.image

        except URLError:
            exit("Unable to connect to server")

    def update_board(self, board):
        try:
            # Attempt to create an image for the updated board
            board = {
                "BoardPositions": board["Board"],
                "HeadColour": self.conf.get("HeadColour", [
                    0,
                    0,
                    255,
                ]),
                "BodyColour": self.conf.get("BodyColour", [
                    50,
                    130,
                    170,
                ]),
                "AppleColour": self.conf.get("AppleColour", [
                    255,
                    0,
                    0,
                ]),
                "BackgroundColour": self.conf.get("BackgroundColour", [
                    60,
                    170,
                    50
                ]),
                "BlockHeight": self.conf.get("BlockHeight", 50),
                "BlockWidth": self.conf.get("BlockWidth", 50),
                "BorderColour": self.conf.get("BorderColour", [
                    255,
                    255,
                    255
                ])
            }

            # Make request
            res = loads(
                urlopen(
                    self.url + "/image",
                    data=dumps(board).encode()
                ).read().decode()
            )

            if "Error" in res.keys():
                # Error creating image, use previous
                print("[WARN]", res.get("Error"))
                return

        except URLError:
            print("[WARN] Unable to connect to server to create image")
            return

        self.image = tk.PhotoImage(data=b64decode(res["Image"]))
        self.label.configure(image=self.image)

    def emit_won(self, res):
        self.lock = True
        # Stop any further requests to server
        w = tk.Label(
            self,
            text="You Won! Final Length " + str(res["Length"]),
            anchor="n",
            height=20
        )
        w.pack()
        
    def emit_dead(self, res):
        self.lock = True
        # Stop any further requests to server
        d = tk.Label(
            self,
            text="You Died! Final Length " + str(res["Length"]),
            anchor="n",
            height=20
        )
        d.pack()
        

if __name__ == "__main__":
    with open("config.json") as c:
        con = load(c)
        # Load from config.json into dict
    app = GUI(con)