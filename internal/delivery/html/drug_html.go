package html

const DrugIndexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Image Uploader</title>

    <style>
    body {
      font-family: "Arial", sans-serif;
      background-color: #f7f7f7;
      color: #333;
      line-height: 1.6;
      padding: 20px;
    }

    header,
    form,
    #selectedFiles,
    button {
      margin-bottom: 20px;
      background: #fff;
      padding: 20px;
      border-radius: 5px;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    }

    h1,
    h3 {
      color: #333;
    }

    /* Form element styling */
    label {
      font-weight: bold;
      display: block;
      margin-bottom: 5px;
    }

    input[type="text"],
    input[type="file"] {
      width: 100%;
      padding: 10px;
      margin-bottom: 10px;
      border: 1px solid #ccc;
      border-radius: 3px;
    }

    input[type="file"] {
      border: none;
    }

    button {
      display: inline-block;
      background: #5cb85c;
      color: #fff;
      padding: 10px 20px;
      border: none;
      border-radius: 5px;
      cursor: pointer;
    }

    button:hover {
      background: #4cae4c;
    }

    button:focus {
      outline: none;
    }

    /* Layout adjustments */
    br {
      display: none;
    }

    /* Add responsiveness to the form */
    @media (max-width: 768px) {
      input[type="text"],
      input[type="file"] {
        width: calc(100% - 20px);
      }

      button {
        width: 100%;
        padding: 15px;
      }
    }

    /* JavaScript-added file inputs styling */
    .new-file-input {
      margin-top: 10px;
    }
  </style>
</head>
<body>
    <header>
        <h1>Image Uploader</h1>
    </header>

    <form action="https://open-data.up.railway.app/api/v1/save/drugs/upload" method="POST" enctype="multipart/form-data">
        <label for="name">Name:</label>
        <input type="text" name="nameOfDrug" required><br>
		<br>
        <label for="description">Description:</label>
        <input type="text" name="description" required><br>
		<br>
        <label for="manufacturer">Manufacturer:</label>
        <input type="text" name="manufacturer" required><br>
		<br>
        <label for="type">type:</label>
        <input type="text" name="type" required><br>
		<br>
		<label for="reciept">reciept:</label>
        <input type="text" name="reciept" required><br>
		<br>
        <label for="images">Images:</label>
        <input type="file" name="images" accept="image/*" multiple required onchange="displaySelectedFiles(this)">
		<br>
        <button type="submit">Upload</button>
    </form>

    <h3>Selected Files:</h3>
    <div id="selectedFiles"></div>

    <button onclick="addFileInput()">Add More Images</button>

   <script>
    function displaySelectedFiles(input) {
        var selectedFilesDiv = document.getElementById("selectedFiles");

        // Log information about name, manufacturer, and description
        console.log("Name: " + document.querySelector('input[name="name"]').value);
        console.log("Manufacturer: " + document.querySelector('input[name="manufacturer"]').value);
        console.log("Description: " + document.querySelector('input[name="description"]').value);

        for (var i = 0; i < input.files.length; i++) {
            var file = input.files[i];
            var img = document.createElement("img");
            img.src = URL.createObjectURL(file);
            selectedFilesDiv.appendChild(img);

            // Log information about each selected file
            console.log("File Name: " + file.name);
            console.log("File Type: " + file.type);
            console.log("File Size: " + file.size + " bytes");
            console.log("Last Modified: " + file.lastModifiedDate);
        }
    }

    function addFileInput() {
        var fileInput = document.createElement("input");
        fileInput.type = "file";
        fileInput.name = "images";
        fileInput.accept = "image/*";
        fileInput.multiple = true;
        fileInput.required = true;
        fileInput.onchange = function () {
            displaySelectedFiles(this);
        };

        document.querySelector("form").appendChild(fileInput);
    }
</script>

</body>
</html>

`

