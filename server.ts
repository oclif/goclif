import * as heroku from 'heroku'
import * as net from 'net'
import * as os from 'os'
import * as path from 'path'
import {inspect} from 'util'

const socket = path.join(os.tmpdir(), 'goclif.sock')

// sets up stdout/stderr which allow us to write the the real stdout
// when running commands it will be mocked out
const stdoutWrite = process.stdout.write
const stderrWrite = process.stdout.write
const stdout = (msg: string) => stdoutWrite.bind(process.stdout)(msg)
const stderr = (msg: string) => stderrWrite.bind(process.stderr)(msg)
const debug = (msg: string) => stderr('server ' + msg + '\n')

type Message = {
  type: 'command'
  argv: string[]
}

function pipeStream(stream: typeof process.stdout, fn: (d: string) => any) {
  stream.write = fn
}

const server = net.createServer()
server.listen(socket, () => {
  debug(`listening: ${socket}`)

  // send the path of the socket back to the go client
  stdout(socket + '\n')
})

server.on('connection', socket => {
  debug('socket connected')
  socket.on('data', data => {
    const message: Message = JSON.parse(data as any)
    debug(`received: ${inspect(message)}`)
    const send = (msg: string) => {
      debug(`sent: ${inspect(msg)}`)
      socket.write(msg)
    }
    const end = () => {
      debug('closing socket\n')
      socket.end()
    }
    pipeStream(process.stdout, send)
    pipeStream(process.stderr, send)

    // run the command
    heroku
      .run(message.argv)
      .then(end)
      .catch(err => {
        if (err.code === 'EEXIT') {
          send(`EEXIT: ${err.oclif.exit}`)
          end()
        } else {
          stderr(err.message)
        }
      })
  })
})

server.on('close', () => {
  debug('closed')
  process.exit(0)
})

setTimeout(() => {
  debug('timed out')
  server.close()
}, 5000)
