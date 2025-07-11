import { render } from 'preact';
import { useState } from 'preact/hooks';
import Header from 'component';

function Counter() {
    const [value, setValue] = useState(0);
    return (
        <>
            <Header />
            <div id="report-1">{value}</div>
            <button id="pinc" onClick={() => setValue(value + 1)}>Increment</button>
            <button id="pdec"  onClick={() => setValue(value - 1)}>Decrement</button>
        </>
    );
}

export function init(e: Element) {
    render(<Counter />, e);
};

