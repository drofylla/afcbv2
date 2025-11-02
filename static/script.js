function copyEmail(elementID) {
  const element = document.getElementById(elementID);
  if (!element) return;

  //copy email
  const copyText = element.innerText;

  navigator.clipboard
    .writeText(copyText)
    .then(() => {
      console.log("Email copied");
    })
    .catch((err) => {
      console.error("Failed to copy email: ", err);
    });
}

// Download contact PDF
function downloadContactPDF(contactId) {
  // Show loading state on the button
  const button = event.target.closest("button");
  const originalHTML = button.innerHTML;
  button.innerHTML =
    '<div class="animate-spin rounded-full h-5 w-5 border-b-2 border-green-500"></div>';
  button.disabled = true;

  // Create download link
  const downloadUrl = `/contacts/${contactId}/pdf`;

  // Create hidden iframe for download
  const iframe = document.createElement("iframe");
  iframe.style.display = "none";
  iframe.src = downloadUrl;
  document.body.appendChild(iframe);

  // Reset button after a short delay
  setTimeout(() => {
    button.innerHTML = originalHTML;
    button.disabled = false;
    document.body.removeChild(iframe);
  }, 2000);
}
