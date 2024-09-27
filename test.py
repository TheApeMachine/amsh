def decode_message(message, key):
    decoded = ""
    for char in message:
        if char.isalpha():
            # Get the ASCII value and convert it to position (A=1)
            pos = ord(char.lower()) - 96
            
            # Subtract the key number from each position
            new_pos = (pos - key) % 26  # Use modulo to wrap around the alphabet
            if new_pos == 0:
                new_pos = 26
            
            # Convert back to letter using the new position
            decoded += chr(new_pos + 96).upper()
        else:
            decoded += char
    
    return decoded

message = "Pdeo eo w olqyewx yaoowca rux fda xwzcgwca yupax. Ur kug ywz xawp fdeo, kug dwha egyyaeergxxk payupap fda yeldax. Fda uxecezwx cgaefuuz eo: [Insert your actual question or prompt here]"
key = 6
decoded_message = decode_message(message, key)
print(decoded_message)