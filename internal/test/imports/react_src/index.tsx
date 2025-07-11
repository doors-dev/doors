import { useState } from "react";
import { createRoot } from "react-dom/client";
import Header from "./component";

function Counter() {
    const [value, setValue] = useState(0);
    return (
        <>
            <Header />
            <div id="report-0">{value}</div>
            <button id="inc"  onClick={() => setValue(value + 1)}>Increment</button>
            <button id="dec" onClick={() => setValue(value - 1)}>Decrement</button>
        </>
    );
}

export function init(e: Element) {
    const root = createRoot(e);
    root.render(<Counter />);
}

