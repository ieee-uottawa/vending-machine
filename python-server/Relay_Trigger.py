import RPi.GPIO as GPIO
import time

# Mapping relay numbers to GPIO pins
RELAY_PINS = {
    1: 2,   2: 3,   3: 4,   4: 17,
    5: 27,  6: 22,  7: 10,  8: 9,
    9: 11, 10: 5, 11: 6, 12: 13,
   13: 19, 14: 26, 15: 14, 16: 15
}

GPIO.setmode(GPIO.BCM)

# Setup all relay pins as outputs and turn them OFF initially
for pin in RELAY_PINS.values():
    GPIO.setup(pin, GPIO.OUT)
    GPIO.output(pin, GPIO.HIGH)  # OFF for active LOW relays

try:
    while True:
        input_string = input("\nEnter relay numbers (1-16) to turn ON (or type 'end' to quit): ")

        if input_string.strip().lower() == "end":
            print("Exiting...")
            break

        try:
            relay_nums = list(map(int, input_string.strip().split()))
        except ValueError:
            print("Invalid input. Please enter numbers like: 1 2 5 12")
            continue

        # Turn selected relays ON
        for num in relay_nums:
            if 1 <= num <= 16:
                GPIO.output(RELAY_PINS[num], GPIO.LOW)  # ON
                print(f"Relay {num} ON")
            else:
                print(f"Relay {num} is out of range (1-16)")

        time.sleep(5)

        # Turn them OFF
        for num in relay_nums:
            if 1 <= num <= 16:
                GPIO.output(RELAY_PINS[num], GPIO.HIGH)  # OFF
                print(f"Relay {num} OFF")

except KeyboardInterrupt:
    print("\nInterrupted by user.")

finally:
    GPIO.cleanup()
    print("GPIO cleaned up.")
