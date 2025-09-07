import RPi.GPIO as GPIO
import time
from typing import List

# Relay number to GPIO pin mapping
RELAY_PINS = {
    1: 2,   2: 3,   3: 4,   4: 17,
    5: 27,  6: 22,  7: 10,  8: 9,
    9: 11, 10: 5, 11: 6, 12: 13,
   13: 19, 14: 26, 15: 14, 16: 15
}

# Slot name to relay number mapping
SLOT_TRIGGERS = {
    "A1": [3, 12, 13, 14],
    "A2": [3, 7, 13, 14],
    "A3": [3, 7, 12, 14],
    "A4": [3, 7, 12, 13], 

    "B1": [2, 12, 13, 14],
    "B2": [2, 7, 13, 14],
    "B3": [2, 7, 12, 14],
    "B4": [2, 7, 12, 13],

    "C1": [5, 12, 13, 14],
    "C2": [5, 7, 13, 14],
    "C3": [5, 7, 12, 14],
    "C4": [5, 7, 12, 13],

    "D1": [4, 16, 15, 14, 13, 12, 10, 8],
    "D2": [4, 16, 15, 14, 13, 10, 8, 7],
    "D3": [4, 16, 15, 14, 12, 10, 8, 7],
    "D4": [4, 16, 15, 13, 12, 10, 8, 7],
    "D5": [4, 16, 14, 13, 12, 7, 8, 10],
    "D6": [4, 16, 14, 13, 12, 7, 8, 15],
    "D7": [4, 15, 14, 13, 12, 10, 8, 7],
    "D8": [4, 16, 15, 14, 13, 12, 10, 7],

    "E1": [1, 16, 15, 14, 13, 12, 10, 8],
    "E2": [1, 16, 15, 14, 13, 10, 8, 7],
    "E3": [1, 16, 15, 14, 12, 10, 8, 7],
    "E4": [1, 16, 15, 13, 12, 10, 8, 7],
    "E5": [1, 16, 14, 13, 12, 7, 8, 10],
    "E6": [1, 16, 14, 13, 12, 7, 8, 15],
    "E7": [1, 15, 14, 13, 12, 10, 8, 7],
    "E8": [1, 16, 15, 14, 13, 12, 10, 7],

    "F1": [6, 12, 13, 14],
    "F2": [6, 7, 13, 14],
    "F3": [6, 7, 12, 14],
    "F4": [6, 7, 12, 13]
}

def activate_slot(relays:List[int]):
    for r in relays:
        GPIO.output(RELAY_PINS[r], GPIO.LOW)
        print(f"Relay {r} ON")
    print()

    time.sleep(2)

    for r in relays:
        GPIO.output(RELAY_PINS[r], GPIO.HIGH)
        print(f"Relay {r} OFF")
    print()

GPIO.setmode(GPIO.BCM)

# Setup all relay pins
for pin in RELAY_PINS.values():
    GPIO.setup(pin, GPIO.OUT)
    GPIO.output(pin, GPIO.HIGH)  # OFF (active LOW)

try:
    while True:
        slot = input("\nEnter slot code (A1 to F4) or 'all' to test every slot sequentially or 'end' to quit: ").strip().upper()

        if slot == "END":
            print("Exiting...")
            break

        print(slot)

        if slot not in SLOT_TRIGGERS and slot != "ALL":
            print("Invalid slot. Try A1 to F4.")
            continue

        if slot == "ALL":
            for s in SLOT_TRIGGERS.keys():
                relays = SLOT_TRIGGERS[s]
                print(f"Activating {s}: relays {relays}\n")
                activate_slot(relays)
            continue

        relays = SLOT_TRIGGERS[slot]
        print(f"Activating {slot}: relays {relays}\n")

        activate_slot(relays)



except KeyboardInterrupt:
    print("\nInterrupted by user.")

finally:
    GPIO.cleanup()
    print("GPIO cleaned up.")
