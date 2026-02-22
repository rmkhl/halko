import React, { useEffect, useState, useRef } from "react";
import { ExecutionChart } from "./ExecutionChart";
import { getApiEndpoints } from "../config/api";

interface LiveExecutionChartProps {
  title?: string;
}

export const LiveExecutionChart: React.FC<LiveExecutionChartProps> = ({
  title = "Live Program Execution"
}) => {
  const [csvData, setCsvData] = useState<string>("");
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    const endpoints = getApiEndpoints();
    const baseUrl = endpoints.controlunit;

    const wsUrl = baseUrl.replace(/^http/, "ws") + "/engine/running/logws";

    const connectWebSocket = () => {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setCsvData(""); // Reset data on new connection
      };

      ws.onmessage = (event) => {
        const message = event.data.trim();

        // Check if message is an error (not CSV data)
        if (message === "No program running") {
          setIsConnected(false);
          setCsvData("");
          return;
        }

        // Skip empty messages
        if (!message) {
          return;
        }

        console.log("Received WebSocket message:", message);

        // Append new CSV line to existing data
        setCsvData((prevData) => {
          if (!prevData) {
            console.log("First message (header):", message);
            return message; // First message is the header
          }
          // Add newline only if previous data doesn't end with one
          const separator = prevData.endsWith("\n") ? "" : "\n";
          const newData = prevData + separator + message;
          console.log("Accumulated CSV lines:", newData.split("\n").length);
          return newData;
        });
      };

      ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        setIsConnected(false);
      };

      ws.onclose = () => {
        setIsConnected(false);
        // Attempt to reconnect after 5 seconds
        setTimeout(connectWebSocket, 5000);
      };
    };

    connectWebSocket();

    // Cleanup on unmount
    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  return (
    <ExecutionChart
      csvData={csvData || undefined}
      title={title}
      isLoading={!csvData && isConnected}
    />
  );
