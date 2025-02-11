const { proxyRouterUrl, agentUsername, agentPassword, modelId } = require('./config');

const basicAuth = Buffer.from(`${agentUsername}:${agentPassword}`).toString("base64");

const tools = [
  {
    type: "function",
    function: {
      name: "get_local_models",
      description: "Returns list of available local AI models to the user",
    },
  },
];

const availableFunctions = {
  "get_local_models": async () => {
    const result = await fetch(`${proxyRouterUrl}/v1/models`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Basic ${basicAuth}`,
      },
    });
    const json = await result.json();
    return JSON.stringify(json);
  },
};

const sendPrompt = (body) => {
  return fetch(`${proxyRouterUrl}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Basic ${basicAuth}`,
      model_id: modelId,
    },
    body: JSON.stringify(body),
  });
};


(async () => {
  const messages = [];
  messages.push({
    role: "user",
    content: "Hi! Can you advise which local model I should use to generate jokes?",
  })

  console.log("Sending prompt: ", messages[0].content);
  const response = await sendPrompt({
    messages,
    tools,
    stream: false,
  });

  const dataRaw = await response.text();
  const data = JSON.parse(dataRaw.replace("data:", ""));

  const choice = data.choices[0].message.tool_calls;
  if (Array.isArray(choice)) {
    for (const toolCall of choice) {
      const functionToCall = availableFunctions[toolCall.function.name];
      if (functionToCall) {
        console.log("Calling function:", toolCall.function.name, "with arguments:", toolCall.function.arguments);
        output = await functionToCall(toolCall.function.arguments);
        console.log("Received function output");

        // Add the function response to messages for the model to use
        messages.push(data.choices[0].message);
        messages.push({
          role: "tool",
          content: output.toString(),
          tool_call_id: toolCall.id,
        });
      } else {
        console.log("Function", tool.function.name, "not found");
      }
    }
  }

  console.log("Sending prompt with function output");
  const response2 = await sendPrompt({
    messages,
    tools,
    stream: false,
  });

  const dataRaw2 = await response2.text();
  const data2 = JSON.parse(dataRaw2.replace("data:", ""));

  console.log("Received response from model");
  console.log(data2.choices[0].message.content);
})();
