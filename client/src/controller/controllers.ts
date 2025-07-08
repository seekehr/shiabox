export const BACKEND_URL = "http://localhost:1323";

export type BadResponse = {
    message: string;
    status_code: number;
};

export type GoodResponse = {
    message: string;
    data: string;
    status_code: number;
};

export type AIMessage = {
    role: string
    content: string
}

export type AIChoice = {
    index: number
    finish_reason: string
    delta: AIMessage
}

export type AIResponse = {
    choices: AIChoice[];
};

export type StreamResponse = {
    message: string;
    data: AIResponse;
    done: boolean;
};

export const handleResponse = async (
    response: Response
): Promise<GoodResponse | BadResponse> => {
    const data = await response.json();

    if (!response.ok) {
        return {
            message: data.message,
            status_code: response.status,
        } as BadResponse;
    }

    return {
        message: data.message,
        data: data.data,
        status_code: response.status,
    } as GoodResponse;
};

export const handleStreamResponse = async (
    response: Response,
    onReceive: (streamResponse: StreamResponse) => void,
    onError: (error: string) => void,
): Promise<void> => {
    try {
        if (!response.ok) {
            const error = await response.json().catch(() => ({
                message: `HTTP error! status: ${response.status}`,
            }));
            onError(error.message);
            return;
        }

        if (!response.body) {
            onError("Response body is missing.");
            return;
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        // keeps running in the background until the stream is done
        (async () => {
            try {
                let unfinished = true;
                while (unfinished) {
                    // read from our SSE stream
                    const { done, value } = await reader.read();
                    if (done) {
                        unfinished = false;
                        break;
                    }


                    // decode the data and chunk it properly to add into buffer
                    buffer += decoder.decode(value, { stream: true });
                    const messages = buffer.split("\n\n");
                    buffer = messages.pop() || "";

                    for (const message of messages) {
                        if (message.startsWith("data: ")) {
                            // remove the data: prefix
                            const jsonStr = message.substring(6);
                            if (jsonStr) {
                                const streamResponse: StreamResponse =
                                    JSON.parse(jsonStr);

                                // this implies an error from the server
                                if (streamResponse.message !== "") {
                                    onError(streamResponse.message);
                                    await reader.cancel();
                                    return;
                                }

                                // handle the stream response
                                onReceive(streamResponse);

                                if (streamResponse.done) {
                                    unfinished = false;
                                    // WE DONE. break loop and cancel the reader
                                    await reader.cancel();
                                    return;
                                }
                            }
                        }
                    }
                }
            } catch (error) {
                if (error instanceof Error && error.name !== "AbortError") {
                    onError("Stream error");
                }
            }
        })();
    } catch (error) {
        if (error instanceof Error && error.name !== 'AbortError') {
            onError(error.message);
        } else if (typeof error !== 'object' || (error && 'name' in error && error.name !== 'AbortError')) {
            onError("An unknown error occurred");
        }
    }
};