/**
 * Pick local infrastructure files using the File System Access API when available,
 * otherwise &lt;input type="file" multiple&gt;.
 */
export function isFileSystemAccessSupported(): boolean {
  return typeof window !== 'undefined' && 'showOpenFilePicker' in window
}

export async function pickInfrastructureFiles(): Promise<File[]> {
  if (typeof window === 'undefined') {
    return []
  }
  if ('showOpenFilePicker' in window) {
    try {
      const handles = await (
        window as unknown as {
          showOpenFilePicker: (opts?: {
            multiple?: boolean
            types?: { description: string; accept: Record<string, string[]> }[]
          }) => Promise<FileSystemFileHandle[]>
        }
      ).showOpenFilePicker({
        multiple: true,
        types: [
          {
            description: 'Terraform state & inventories',
            accept: {
              'application/json': ['.json', '.tfstate'],
              'text/plain': ['.ini', '.tfstate', '.yaml', '.yml'],
            },
          },
        ],
      })
      const out: File[] = []
      for (const h of handles) {
        out.push(await h.getFile())
      }
      return out
    } catch (e) {
      const err = e as { name?: string }
      if (err?.name === 'AbortError') {
        return []
      }
      throw e
    }
  }
  return new Promise((resolve, reject) => {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = true
    input.accept = '.json,.tfstate,.ini,.yaml,.yml,text/plain,application/json'
    input.onchange = () => {
      const list = input.files ? Array.from(input.files) : []
      resolve(list)
    }
    input.click()
    input.addEventListener(
      'error',
      () => {
        reject(new Error('file picker failed'))
      },
      { once: true },
    )
  })
}

function contentTypeForFileName(name: string): string {
  const lower = name.toLowerCase()
  if (lower.endsWith('.tfstate') || lower.endsWith('.json')) {
    return 'application/json'
  }
  if (lower.endsWith('.ini') || lower.endsWith('.txt')) {
    return 'text/plain'
  }
  if (lower.endsWith('.yaml') || lower.endsWith('.yml')) {
    return 'application/yaml'
  }
  return 'application/octet-stream'
}

export function fileToIngestPayload(file: File): {
  name: string
  contentType: string
  encoding: string
  data: string
  clientPathHint?: string
  lastModified?: string
} {
  const lastMod =
    file.lastModified && !Number.isNaN(file.lastModified)
      ? new Date(file.lastModified).toISOString()
      : undefined
  return {
    name: file.name,
    contentType: contentTypeForFileName(file.name),
    encoding: 'utf8',
    data: '', // filled after readAsText
    clientPathHint: file.webkitRelativePath && file.webkitRelativePath !== file.name ? file.webkitRelativePath : file.name,
    lastModified: lastMod,
  }
}

export async function readFilesForIngest(
  files: File[],
): Promise<
  {
    name: string
    contentType: string
    encoding: string
    data: string
    clientPathHint?: string
    lastModified?: string
  }[]
> {
  const out: {
    name: string
    contentType: string
    encoding: string
    data: string
    clientPathHint?: string
    lastModified?: string
  }[] = []
  for (const f of files) {
    const base = fileToIngestPayload(f)
    const text = await f.text()
    out.push({ ...base, data: text })
  }
  return out
}
