const net = require('net');

const server = net.createServer((socket) => {
  socket.on('data', (data) => {
    socket.write(data);
  });
  
  socket.on('error', (err) => {
    // ignore
  });
});

const port = 1514;
server.listen(port, () => {
  console.log(`Node server listening on :${port}`);
});
