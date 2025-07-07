export const BACKEND_URL = "http://localhost:1323";

export type BackendBadResponse = {
    message: string;
    status_code: string;
};

export type BackendGoodResponse = {
    message: string;
    data: string;
    status_code: string;
};

export const handleResponse = async (response: Response): Promise<BackendGoodResponse|BackendBadResponse> => {
    const data = await response.json();
    if (!response.ok) {
        return {
            message: data.message,
            status_code: data.status_code,
        } as BackendBadResponse;
    }

    return {
        message: data.message,
        data: data.data,
        status_code: data.status_code,
    } as BackendGoodResponse;
};