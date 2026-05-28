const sessionInvalidMessages = [
  "unauthenticated",
  "invalid token",
  "token is revoked",
  "user not found",
];

// Сессия считается протухшей только для ошибок, которые требуют нового входа.
export function isSessionInvalidError(error: unknown) {
  return errorMessages(error).some((message) =>
    sessionInvalidMessages.some((invalidMessage) =>
      message.toLowerCase().includes(invalidMessage),
    ),
  );
}

function errorMessages(error: unknown): string[] {
  if (!isRecord(error)) {
    return [];
  }

  const messages: string[] = [];
  if (typeof error.message === "string") {
    messages.push(error.message);
  }

  const response = error.response;
  if (isRecord(response) && Array.isArray(response.errors)) {
    for (const graphQLError of response.errors) {
      if (isRecord(graphQLError) && typeof graphQLError.message === "string") {
        messages.push(graphQLError.message);
      }
    }
  }

  return messages;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
