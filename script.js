const consoleDiv = document.getElementById('console');
const serialDataDiv = document.getElementById('serial-data');
const socket = new WebSocket('ws://localhost:3000/data');

socket.addEventListener('message', (event) => {
  const data = event.data;
  console.log(`Received data from serial device: ${data}`);
  serialDataDiv.innerHTML += `<p>${data}</p>`;
  serialDataDiv.scrollTop = serialDataDiv.scrollHeight;
});

async function connect() {
  try {
    const devices = await navigator.hid.requestDevice({ filters: [{ usagePage: 0x01 }] });
    const device = devices[0];
    await device.open();
    consoleDiv.innerHTML += `<p class="timestamp">${new Date().toLocaleString()}</p><p class="input">Connected to device: ${device.productName}</p>`;
    device.addEventListener('inputreport', (event) => {
      consoleDiv.innerHTML += `<p class="timestamp">${new Date().toLocaleString()}</p><p class="input">${event.data}</p>`;
    });
  } catch (error) {
    console.log(error);
  }
}

connect();
