import { Binary } from "bson"

const base64Digits = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

const HexToBase64 = (hex: string) => {
    let base64 = ""
    let group: number
    for (let i = 0; i < 30; i += 6) {
        group = parseInt(hex.substr(i, 6), 16)
        base64 += base64Digits[(group >> 18) & 0x3f]
        base64 += base64Digits[(group >> 12) & 0x3f]
        base64 += base64Digits[(group >> 6) & 0x3f]
        base64 += base64Digits[group & 0x3f]
    }
    group = parseInt(hex.substr(30, 2), 16)
    base64 += base64Digits[(group >> 2) & 0x3f]
    base64 += base64Digits[(group << 4) & 0x3f]
    base64 += "=="
    return base64
}

export const ToBinary = (uuid: string) => {
    if (!uuid || typeof uuid !== "string") {
        throw new Error("Invalid UUID format")
    }

    const hex = uuid.replace(/[{}-]/g, "")
    if (hex.length !== 32) {
        throw new Error("Invalid UUID format after stripping characters")
    }

    try {
        return Binary.createFromHexString(hex, 3)
    } catch (error: any) {
        console.error('Original UUID:', uuid);
        console.error('Hex:', hex);
        throw new Error("Failed to create BSON Binary from hex: " + error.message)
    }
}

const base64ToHex = (base64: string) => {
    let binaryString = ""
    let hexString = ""

    for (let i = 0; i < base64.length; i += 4) {
        const group = (base64Digits.indexOf(base64[i]) << 18) |
            (base64Digits.indexOf(base64[i + 1]) << 12) |
            (base64Digits.indexOf(base64[i + 2]) << 6) |
            base64Digits.indexOf(base64[i + 3])
        binaryString += String.fromCharCode((group >> 16) & 0xff) +
            String.fromCharCode((group >> 8) & 0xff) +
            String.fromCharCode(group & 0xff)
    }

    for (let i = 0; i < binaryString.length; i++) {
        const hex = binaryString.charCodeAt(i).toString(16)
        hexString += (hex.length === 1 ? "0" : "") + hex
    }

    return hexString
}

export const FromBinary = (binary: any) => {
    // Handle Binary object with buffer
    if (binary?.buffer instanceof Uint8Array) {
        // Convert buffer to hex string without reordering
        const hex = Array.from(binary.buffer)
            .map(b => b.toString(16).padStart(2, '0'))
            .join('');

        // Format as UUID
        return [
            hex.substr(0, 8),
            hex.substr(8, 4),
            hex.substr(12, 4),
            hex.substr(16, 4),
            hex.substr(20, 12)
        ].join('-');
    }

    return binary.toString();
}