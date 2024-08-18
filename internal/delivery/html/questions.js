const questions=localStorage.getItem('numberquestion')
const container = document.querySelector('.container');
const fact_id=localStorage.getItem('factId')

for (let i = 0; i < Number(questions); i++) {
    // Create a div for each question
    const questionDiv = document.createElement("div");
    questionDiv.className = "fact";
    questionDiv.id = `question-${i}`;
    questionDiv.innerHTML = `
        <div class="fact">
            <label>Question ${i + 1}</label>
            <input type="text" name="question-${i}">
        </div>
    `;

    // Create a container for the choices related to this question
    const choicesContainer = document.createElement('div');
    choicesContainer.className = 'choices-container';

    // Create the choices with checkboxes
    for (let j = 0; j < 3; j++) {
        const choiceDiv = document.createElement('div');
        choiceDiv.className = 'fact';
        choiceDiv.id = `choice-${i}-${j}`;
        choiceDiv.innerHTML = `
            <div class="fact">
                <label>Choice ${j + 1}</label>
                <input type="text" name="choice-${i}-${j}">
                <input type="checkbox" name="correct-${i}-${j}" value="true"> Correct
            </div>
        `;
        choicesContainer.appendChild(choiceDiv);
    }

    // Append the choices container to the question div
    questionDiv.appendChild(choicesContainer);

    // Append the question div to the main container
    container.appendChild(questionDiv);
}
container.innerHTML +=` <div id="submitBtn" className="next">
    <button id="next" onclick="submitBtn()">
        Next
    </button>
</div>`
// Event listener for submit button
function submitBtn() {
    const questionsData = [];

    for (let i = 0; i < Number(questions); i++) {
        const questionInput = document.querySelector(`input[name="question-${i}"]`).value;
        const choices = [];

        for (let j = 0; j < 3; j++) {
            const choiceInput = document.querySelector(`input[name="choice-${i}-${j}"]`).value;
            const isCorrect = document.querySelector(`input[name="correct-${i}-${j}"]`).checked;
            choices.push({
                content: choiceInput,
                is_true: isCorrect
            });
        }

        questionsData.push({
            fact_id: Number(fact_id),
            question: questionInput,
            choices: choices
        });
    }

    // Convert the questionsData array to JSON
    const jsonData = JSON.stringify(questionsData);
    console.log(jsonData); // This will log the JSON string

    // Send the JSON data via POST request
    fetch('/api/v1/fact/createQuestions', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: jsonData,
    })
        .then(response => response.json())
        .then(data => {
            console.log('Success:', data);
            window.location.href='/api/v1/create/fact';
            // Handle the response from the server
        })
        .catch((error) => {
            console.error('Error:', error);
            // Handle any errors
        });
}
