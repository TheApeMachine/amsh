onmessage = (event) => {
    console.log("worker.js", "onmessage", event);
    const { topic, effect } = event.data;
    postMessage({ topic, effect });
};