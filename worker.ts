import * as heroku from 'heroku'
import * as net from 'net'
import * as path from 'path'
import {inspect} from 'util'

const mockStdin = require('mock-stdin').stdin()

const sockets = {
  ctl: path.join(process.argv[3], 'ctl'),
  stdin: path.join(process.argv[3], 'stdin'),
  stdout: path.join(process.argv[3], 'stdout'),
  stderr: path.join(process.argv[3], 'stderr'),
}

const realExit = process.exit

// sets up stdout/stderr which allow us to write the the real stdout
// when running commands it will be mocked out
const stdoutWrite = process.stdout.write
const stderrWrite = process.stdout.write
const stdout = (msg: string) => stdoutWrite.bind(process.stdout)(msg)
const stderr = (msg: string) => stderrWrite.bind(process.stderr)(msg)
const debug = (msg: string) => process.env.DEBUG ? stderr('worker ' + msg + '\n') : null

type Message = {
  id: string
  worker_id: string
  type: string
  argv: string[]
}

function pipeStream(stream: typeof process.stdout, fn: (d: string) => any) {
  stream.write = fn
}

function openSocket(id: keyof typeof sockets): Promise<net.Server> {
  return new Promise(resolve => {
    const server = net.createServer()
    const socket = sockets[id]
    server.listen(socket, () => {
      debug(`listening: ${socket}`)
      resolve(server)
    })
  })
}

const send = (socket: net.Socket, msg: string) => {
  debug(`sent: ${inspect(msg)}`)
  socket.write(msg)
}

Promise.all([openSocket('ctl'), openSocket('stdin'), openSocket('stdout'), openSocket('stderr')])
  .then(servers => {
    for (let server of servers) {
      server.on('close', () => {
        debug('closed')
        process.exit(0)
      })
    }
    let timeout: NodeJS.Timer
    const resetTimeout = () => {
      timeout = setTimeout(() => {
        debug('timed out')
        for (let server of servers) server.close()
      }, 10000)
    }
    resetTimeout()
    const sockets = {
      ctl: servers[0],
      stdin: servers[1],
      stdout: servers[2],
      stderr: servers[3],
    }
    sockets.stdin.on('connection', socket => {
      debug('stdin socket connected')
      socket.on('data', d => {
        debug(`stdin: ${inspect(d.toString())}`)
        mockStdin.send(d)
      })
    })
    sockets.stdout.on('connection', socket => {
      debug('stdout socket connected')
      pipeStream(process.stdout, msg => {
        debug(`stdout: ${inspect(msg.toString())}`)
        socket.write(msg)
      })
    })
    sockets.stderr.on('connection', socket => {
      debug('stderr socket connected')
      pipeStream(process.stderr, msg => {
        debug(`stderr: ${inspect(msg.toString())}`)
        socket.write(msg)
      })
    })
    let ctlSockets: net.Socket[] = []
    sockets.ctl.on('connection', socket => {
      clearTimeout(timeout)
      ctlSockets.push(socket)
      socket.on('close', () => {
        ctlSockets = ctlSockets.filter(c => c !== socket)
      })
      socket.on('data', data => {
        const message: Message = JSON.parse(data as any)
        debug(`received: ${inspect(message)}`)
        const end = () => {
          debug('closing socket\n')
          socket.end()
        }
        (process.exit as any) = exit
        heroku
          .run(message.argv)
          .then(() => {
            exit(0)
          })
          .catch(err => {
            if (err.code === 'EEXIT') {
              exit(err.oclif.exit)
              end()
            } else {
              process.stderr.write(err.stack)
              exit(1)
            }
          })
      })
    })

    function exit(code: number) {
      for (let socket of ctlSockets) {
        send(socket, JSON.stringify({code}))
      }
      resetTimeout()
      process.exit = realExit
    }
    stdout('up\n')
  })
