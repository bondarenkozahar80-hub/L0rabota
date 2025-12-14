function getOrder() {
    const orderId = document.getElementById('orderId').value.trim();
    const errorDiv = document.getElementById('error');
    const loadingDiv = document.getElementById('loading');
    const orderInfoDiv = document.getElementById('orderInfo');

    errorDiv.style.display = 'none';
    orderInfoDiv.style.display = 'none';
    loadingDiv.style.display = 'block';

    if (!orderId) {
        showError('Please enter an Order ID');
        return;
    }

    fetch(`/order/${orderId}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Order not found');
            }
            return response.json();
        })
        .then(order => {
            displayOrder(order);
        })
        .catch(error => {
            showError(error.message);
        })
        .finally(() => {
            loadingDiv.style.display = 'none';
        });
}

function displayOrder(order) {
    document.getElementById('orderBasic').innerHTML = `
        <p><strong>Order UID:</strong> ${order.order_uid}</p>
        <p><strong>Track Number:</strong> ${order.track_number}</p>
        <p><strong>Entry:</strong> ${order.entry}</p>
        <p><strong>Customer ID:</strong> ${order.customer_id}</p>
        <p><strong>Delivery Service:</strong> ${order.delivery_service}</p>
        <p><strong>Date Created:</strong> ${new Date(order.date_created).toLocaleString()}</p>
    `;

    document.getElementById('deliveryInfo').innerHTML = `
        <p><strong>Name:</strong> ${order.delivery.name}</p>
        <p><strong>Phone:</strong> ${order.delivery.phone}</p>
        <p><strong>Email:</strong> ${order.delivery.email}</p>
        <p><strong>Address:</strong> ${order.delivery.city}, ${order.delivery.address}, ${order.delivery.region} ${order.delivery.zip}</p>
    `;

    document.getElementById('paymentInfo').innerHTML = `
        <p><strong>Transaction:</strong> ${order.payment.transaction}</p>
        <p><strong>Amount:</strong> $${(order.payment.amount / 100).toFixed(2)}</p>
        <p><strong>Currency:</strong> ${order.payment.currency}</p>
        <p><strong>Provider:</strong> ${order.payment.provider}</p>
        <p><strong>Bank:</strong> ${order.payment.bank}</p>
        <p><strong>Payment Date:</strong> ${new Date(order.payment.payment_dt * 1000).toLocaleString()}</p>
    `;

    const itemsHtml = order.items.map(item => `
        <div class="item">
            <p><strong>Name:</strong> ${item.name}</p>
            <p><strong>Brand:</strong> ${item.brand}</p>
            <p><strong>Price:</strong> $${(item.price / 100).toFixed(2)}</p>
            <p><strong>Total Price:</strong> $${(item.total_price / 100).toFixed(2)}</p>
            <p><strong>Sale:</strong> ${item.sale}%</p>
            <p><strong>Status:</strong> ${item.status}</p>
        </div>
    `).join('');

    document.getElementById('itemsInfo').innerHTML = itemsHtml;
    document.getElementById('orderInfo').style.display = 'block';
}

function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

document.getElementById('orderId').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        getOrder();
    }
});
