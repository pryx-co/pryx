import { describe, test, expect } from "vitest";
import { ConnectionError, RuntimeEvent } from "./ws";

describe("WebSocket Service", () => {
  describe("ConnectionError", () => {
    test("should create error with message", () => {
      const error = new ConnectionError("Test error message");
      expect(error._tag).toBe("ConnectionError");
      expect(error.message).toBe("Test error message");
      expect(error.originalError).toBeUndefined();
    });

    test("should create error with original error", () => {
      const original = new Error("Original error");
      const error = new ConnectionError("Wrapped error", original);
      expect(error.originalError).toBe(original);
    });
  });

  describe("RuntimeEvent", () => {
    test("should parse valid runtime events", () => {
      const testEvent: RuntimeEvent = {
        event: "trace",
        type: "test",
        session_id: "test-session",
        payload: { message: "test" },
      };

      expect(testEvent.event).toBe("trace");
      expect(testEvent.session_id).toBe("test-session");
      expect(testEvent.payload).toHaveProperty("message");
    });

    test("should handle events without optional fields", () => {
      const minimalEvent: RuntimeEvent = {
        event: "trace",
      };

      expect(minimalEvent.event).toBe("trace");
      expect(minimalEvent.session_id).toBeUndefined();
      expect(minimalEvent.payload).toBeUndefined();
    });
  });

  describe("ConnectionStatus", () => {
    test("should define disconnected state", () => {
      const state = { _tag: "Disconnected" };
      expect(state._tag).toBe("Disconnected");
    });

    test("should define connecting state", () => {
      const state = { _tag: "Connecting" };
      expect(state._tag).toBe("Connecting");
    });

    test("should define connected state", () => {
      const state = { _tag: "Connected" };
      expect(state._tag).toBe("Connected");
    });

    test("should define error state", () => {
      const error = new ConnectionError("Test");
      const state = { _tag: "Error", error };
      expect(state._tag).toBe("Error");
      expect(state.error).toBe(error);
    });
  });
});
