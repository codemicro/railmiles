export const roundFloat = (x, decimalPlaces) => {
    const scale = Math.pow(10, decimalPlaces)
    x *= scale
    x = Math.round(x)
    x /= scale
    return x
}

const baseURL = ""

export const makeURL = (path) => {
    if (baseURL.endsWith("/") && path.startsWith("/")) {
        return baseURL + path.substring(1)
    }
    return baseURL + path
}

export const leftPad = (str, char, len) => {
    str = str.toString()
    if (str.length >= len) {
        return str
    }

    while (str.length < len) {
        str = char + str
    }

    return str
}

const dateFormat = {year: 'numeric', month: 'short', day: 'numeric'};

export const formatDate = (date) => {
    return new Date(Date.parse(date)).toLocaleDateString(undefined, dateFormat)
}