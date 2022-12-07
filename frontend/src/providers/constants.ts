function getBaseUrl(): string {

    let url: string = `${window.location.href}`;
    let hostName: string[] = `${window.location.href}`.split("/");
    let hostNameParts: string[] = url.split(":");
    let baseURL: string;

    // console.log(hostName[2])
    // console.log(window.location.hostname)

    if (hostNameParts.length > 0) {
        hostNameParts[2] === undefined ? (
            baseURL = `${hostNameParts[0]}://${hostName[2]}`
        ) : (
            baseURL = `${hostNameParts[0]}://${hostName[2]}`)
        console.log(baseURL)
        return baseURL
    }
    return `https://${window.location.hostname}`;
}

const isDev = process.env.NODE_ENV;
const REGISTRY_ADMIN_DEV_URL= process.env.REACT_APP_RA_DEV_URL ? (process.env.REACT_APP_RA_DEV_URL as string) : `http://${window.location.hostname}`

export const API_BASE: string = '/api/v1';
export const API_AUTH: string = '/auth';
export const BASE_URL: string = isDev === "development" ? REGISTRY_ADMIN_DEV_URL : getBaseUrl();
export const BASENAME: string = "web";

