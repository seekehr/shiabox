import { BACKEND_URL, handleResponse, type BackendBadResponse, type BackendGoodResponse } from "./controllers";

const sendRequestAIUrl = `${BACKEND_URL}/ai/request`;

const sendRequestAI = async (request: string): Promise<BackendGoodResponse|BackendBadResponse> => {
  const response = await fetch(sendRequestAIUrl, {
    method: "POST",
    headers: {
        "Content-Type": "application/json",
    },
    body: JSON.stringify({ "prompt": request }),
  });
  return await handleResponse(response);
};

export { sendRequestAI };