import OpenAI from "openai";
import readline from "readline";

const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY,
});

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
});

const messages = [
  { role: "system", content: "You are a helpful assistant." }
];

console.log("AI Bot: Type 'exit' to quit.\n");

rl.on("line", async (input) => {
  if (input.toLowerCase() === "exit") {
    rl.close();
    return;
  }

  messages.push({ role: "user", content: input });

  const response = await openai.chat.completions.create({
    model: "gpt-4o-mini",
    messages: messages,
  });

  const reply = response.choices[0].message.content;
  console.log("Bot:", reply);

  messages.push({ role: "assistant", content: reply });
});
