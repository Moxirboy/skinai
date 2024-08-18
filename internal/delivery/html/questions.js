const questions=localStorage.getItem('numberquestion')

const container=document.querySelector('.container')
console.log(questions)

for (let i=0;i<Number(questions);i++){
    const questionDiv=document.createElement("div")
    questionDiv.className="fact"
    questionDiv.id=`question-${i}`
    questionDiv.innerHTML=`
        <div class="fact">
            <label>question</label>
            <input type="text">
        </div>
`
    container.appendChild(questionDiv)
    for (let j=0;j<3;j++){
        const choiceDiv = document.createElement('div');
        choiceDiv.className = 'fact';
        choiceDiv.id=`choice-${j}`
        choiceDiv.innerHTML= `
        <div class="fact">
            <label>choice</label>
            <input type="text">
        </div>
`
        container.appendChild(choiceDiv)
    }
}
container.innerHTML+=`
<div class="next">
            <button onclick="saveQuestions(event)">
                Next
            </button>
        </div>
`
function SaveQuestions(event){

}