"use client";

import { useEffect, useState } from "react";

export default function Home() {
  const [events, setEvents] = useState<any[]>([]);

  useEffect(() => {
    const streamUrl = process.env.NEXT_PUBLIC_SERVICE_A_STREAM_URL || "http://localhost:8080/stream";
    const evtSource = new EventSource(streamUrl);

    evtSource.onmessage = (e) => {
      const data = JSON.parse(e.data);
      setEvents((prev) => [...prev, data]);
    };

    return () => evtSource.close();
  }, []);

  return (
    <html>
      
      <head>
        <title>Eventi ricevuti</title>
      </head>
      <body>
        <main className="p-4">
          <h1 className="text-2xl font-bold">Eventi ricevuti</h1>
          <ul>
            {events.map((e, i) => (
              <li key={i}>{e.message}</li>
            ))}
          </ul>
        </main>
      </body>
    </html>
  );
}
