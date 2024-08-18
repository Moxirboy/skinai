const questions=localStorage.getItem('numberquestion')

const container=document.querySelector('.questions-container')
console.log(questions)

for (let i=0;i<Number(questions);i++){

    container.innerHTML+= `
        <div class="fact">
            <label>Fact title</label>
            <input type="text">
        </div>

`

}