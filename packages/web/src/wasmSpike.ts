/** Minimal Wasm module exporting add(i32,i32)->i32; proves browser Wasm loading path. */
export const spikeWasmBytes = new Uint8Array([
  0, 97, 115, 109, 1, 0, 0, 0, 1, 7, 1, 96, 2, 127, 127, 1, 127, 3, 2, 1, 0, 7, 7, 1, 3, 97, 100, 100, 0, 0, 10, 9, 1, 7, 0, 32, 0, 32, 1, 106, 11,
])

export async function wasmSpikeAdd(a: number, b: number): Promise<number> {
  const { instance } = await WebAssembly.instantiate(spikeWasmBytes)
  const add = instance.exports.add as (x: number, y: number) => number
  return add(a, b)
}

export function wasmSpikeEnabled(): boolean {
  return import.meta.env.VITE_ENABLE_WASM_SPIKE === 'true'
}
