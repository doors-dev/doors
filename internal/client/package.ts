
const decoder = new TextDecoder()

export type Package = {
    start: number,
    end: number,
    payload: string,
    isFiller: boolean,
    action: string,
    arg: any,
}

export class PackageBuilder {
    private payloadParts: Array<string> = []
    private headerParts: Array<string> = []
    build(): Package {
        const header = JSON.parse(this.headerParts.join(""))
        return {
            end: header[0][0],
            start: header[0].length == 2 ? header[0][1] : header[0][0],
            action: header.length == 2 ? header[1][0] : "",
            arg: header.length == 2 ? header[1][1] : undefined,
            isFiller: header.length == 1,
            payload: this.payloadParts.join(""),
        }
    }
    appendHeaderData(buf: Uint8Array) {
        if (buf.length == 0) {
            return
        }
        this.headerParts.push(decoder.decode(buf))
    }
    appendPayloadData(buf: Uint8Array) {
        if (buf.length == 0) {
            return
        }
        this.payloadParts.push(decoder.decode(buf))
    }
}
