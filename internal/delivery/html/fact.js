// fact.js
const submitBtn = document.getElementById('next');

submitBtn.addEventListener("click", function() {
        const factTitle = document.querySelector(".fact input").value;
        const content = document.querySelector(".content input").value;
        const numberOfQuestions = document.querySelector(".number_question input").value;

        console.log("Fact title:", factTitle);
        console.log("Content:", content);
        console.log("Number of questions:", numberOfQuestions);
        const fact = {
            title: factTitle,
            content: content,
            number_of_question: Number(numberOfQuestions)
        };
        data=JSON.stringify(fact)
    console.log(data)
        fetch('/api/v1/fact/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: data,
        })
            .then(response => response.json())
            .then(data => {
                console.log('Success:', data);
                localStorage.setItem('factId', data.Id);
                localStorage.setItem('numberquestion', numberOfQuestions);
                console.log('ID saved to localStorage:', data.Id);
                window.location.href='/api/v1/create/fact/question';

            })
            .catch((error) => {
                console.error('Error:', error);
            });
        // If you want to do something with these values, you can do it here
        // For example, send them to a server or store them locally

});



