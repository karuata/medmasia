(function(){
  const root=document.getElementById("apex-chatbot");
  if(!root)return;
  const email="rodrigomasini.ai@gmail.com";
  const sessionKey="medmasia_bridgebot_session_id";
  const copy={
    kicker:"Assistente",
    title:"MedMasIA Assistant",
    intro:"Sou o assistente do MedMasIA. Posso ajudar a entender onde a rotina da clínica está vazando contexto: potenciais pacientes sem retorno, documentos espalhados, decisões sem registro e equipe dependente da memória do médico. Para dados clínicos ou casos sensíveis, o melhor caminho é uma conversa reservada.",
    placeholder:"Digite sua pergunta",
    sendLabel:"Enviar",
    closeLabel:"Fechar",
    openLabel:"Abrir assistente",
    fallback:"Não tenho contexto suficiente para responder com precisão. Para um caso específico, recomendo uma conversa inicial para entender sua rotina, equipe, documentos e riscos.",
    fallbackLinks:[{label:"Solicitar conversa",href:"#contact"},{label:"Email",href:`mailto:${email}`}],
    suggestions:[
      "Onde minha rotina está vazando?",
      "O que é o twin?",
      "Já uso IA. O que muda?",
      "Isso substitui decisão clínica?",
      "Integra com agenda ou CRM?",
      "O que fica funcionando?"
    ],
    entries:[
      {
        keys:["vazando","vaza","perdendo","caos","rotina","memoria","memória","whatsapp","gargalo"],
        answer:"Os vazamentos mais comuns são: paciente sem retorno, documento espalhado, decisão sem registro, equipe perguntando a mesma coisa e médico gestor virando sistema operacional da clínica. O MedMasIA começa mapeando um fluxo real antes de implantar qualquer automação.",
        links:[{label:"Ver dor",href:"#problem"},{label:"Diagnóstico",href:"#contact"}]
      },
      {
        keys:["pacote","programa","oferta","inclui","comprando","implantacao","implantação"],
        answer:"MedMasIA é um programa de mentoria e implantação, não apenas uma ferramenta. O pacote combina sessões semanais, configuração guiada de um twin via WhatsApp, BridgeBot de relacionamento com pacientes, base de conhecimento, rotina de decisões e regras de governança clínica.",
        links:[{label:"Ver pacote",href:"#offer"},{label:"Entregáveis",href:"#deliverables"}]
      },
      {
        keys:["digital twin","twin","whatsapp","assistente","sirius","o que e","o que é"],
        answer:"No MedMasIA, twin é uma memória operacional conversável no WhatsApp. Ele não é um clone do médico. Ele ajuda a transformar mensagens, documentos, áudios e decisões soltas em contexto pesquisável, pendências, próximos passos e respostas baseadas no que foi aprovado.",
        links:[{label:"Entender twin",href:"#what-is-twin"},{label:"Produtos",href:"#twin"}]
      },
      {
        keys:["mentoria","aprender","zero","0","sessoes","semanais","ia","nivel","nível","maturidade","iniciante"],
        answer:"A mentoria muda conforme o ponto de partida. Para quem começa, cobre uso seguro e prompts. Para quem já usa IA, transforma improviso em rotina. Para clínicas mais maduras, foca equipe, métricas, governança, relacionamento com pacientes e evolução dos produtos implantados.",
        links:[{label:"Ver níveis",href:"#levels"},{label:"Frentes",href:"#mentoria"}]
      },
      {
        keys:["ja uso","já uso","avancado","avançado","chatgpt","gpt","crm","processo"],
        answer:"Se você já usa IA, o valor não é aprender o básico. É transformar uso individual em capacidade operacional: base aprovada, fluxos com equipe, critérios de handoff, métricas, revisão de risco e produtos implantados na rotina da clínica.",
        links:[{label:"Ver níveis",href:"#levels"},{label:"Entregáveis",href:"#deliverables"}]
      },
      {
        keys:["agenda","agendamento","bridgebot","paciente","pacientes","conversao","crm","relacionamento"],
        answer:"O BridgeBot é o fluxo de relacionamento com potenciais pacientes. Ele pode responder perguntas aprovadas, entender intenção, capturar contexto mínimo, sinalizar handoff humano e preparar o próximo passo para a equipe. Integrações reais de agenda, CRM ou prontuário dependem do escopo técnico da clínica.",
        links:[{label:"Ver oferta",href:"#offer"},{label:"Solicitar conversa",href:"#contact"}]
      },
      {
        keys:["documento","documentos","pdf","exame","contrato","protocolo","prontuario","arquivo","entregavel","entregáveis","funcionando"],
        answer:"A camada de conhecimento pode organizar documentos pessoais, administrativos, organizacionais e médicos quando houver permissão e ambiente adequado. Exemplos: protocolos, contratos, treinamentos, orientações, exames, políticas e materiais da clínica.",
        links:[{label:"Ver entregáveis",href:"#deliverables"},{label:"Solicitar conversa",href:"#contact"}]
      },
      {
        keys:["diagnostico","prescricao","conduta","clinica","substitui","risco","lgpd","privacidade","não faz","nao faz"],
        answer:"MedMasIA não posiciona IA como médico autônomo. Não faz diagnóstico autônomo, prescrição automática ou conduta clínica sem revisão. A IA apoia organização, memória, triagem operacional, decisão assistida e documentação, com governança e handoff humano.",
        links:[{label:"Ver governança",href:"#governanca"},{label:"Solicitar conversa",href:"#contact"}]
      },
      {
        keys:["preco","valor","custa","contratar","disponibilidade","orcamento"],
        answer:"Formato e investimento dependem do escopo: duração do programa, quantidade de sessões, twin, BridgeBot, documentos, equipe, integrações necessárias, suporte e nível de governança. O primeiro passo é um diagnóstico de 30 minutos.",
        links:[{label:"Solicitar conversa",href:"#contact"},{label:"Email",href:`mailto:${email}`}]
      },
      {
        keys:["quem","publico","medico","clinica","dono","gestor","empreendedor"],
        answer:"A oferta foi desenhada para médicos gestores, donos de clínicas, especialistas com alta demanda, empreendedores em saúde e equipes premium que precisam organizar decisões, documentos, pacientes e rotina com IA.",
        links:[{label:"Para quem",href:"#forwhom"},{label:"Solicitar conversa",href:"#contact"}]
      }
    ]
  };
  const normalize=(value)=>String(value||"").toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g,"");

  function renderShell(){
    root.className="apex-chatbot";
    root.innerHTML="";
    const button=document.createElement("button");
    button.className="apex-chatbot__button";
    button.type="button";
    button.setAttribute("aria-label",copy.openLabel);
    button.innerHTML='<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4.5 6.5h15v9h-8l-4.5 3v-3H4.5z"/><path d="M8 10h.01M12 10h.01M16 10h.01"/></svg>';
    const panel=document.createElement("section");
    panel.className="apex-chatbot__panel";
    panel.setAttribute("aria-label",copy.title);
    panel.innerHTML=`
      <div class="apex-chatbot__foil"></div>
      <div class="apex-chatbot__head">
        <div><div class="apex-chatbot__kicker">${copy.kicker}</div><div class="apex-chatbot__title">${copy.title}</div></div>
        <button class="apex-chatbot__close" type="button" aria-label="${copy.closeLabel}">x</button>
      </div>
      <div class="apex-chatbot__messages" role="log" aria-live="polite"></div>
      <div class="apex-chatbot__suggestions"></div>
      <form class="apex-chatbot__form">
        <input class="apex-chatbot__input" type="text" autocomplete="off" placeholder="${copy.placeholder}">
        <button class="apex-chatbot__send" type="submit" aria-label="${copy.sendLabel}">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12h14"/><path d="m13 6 6 6-6 6"/></svg>
        </button>
      </form>`;
    root.append(panel,button);
    button.addEventListener("click",()=>root.classList.toggle("is-open"));
    panel.querySelector(".apex-chatbot__close").addEventListener("click",()=>root.classList.remove("is-open"));
    panel.querySelector(".apex-chatbot__form").addEventListener("submit",(event)=>{
      event.preventDefault();
      const input=panel.querySelector(".apex-chatbot__input");
      ask(input.value);
      input.value="";
    });
    copy.suggestions.forEach((suggestion)=>{
      const chip=document.createElement("button");
      chip.type="button";
      chip.textContent=suggestion;
      chip.addEventListener("click",()=>ask(suggestion));
      panel.querySelector(".apex-chatbot__suggestions").appendChild(chip);
    });
    addMessage(copy.intro,"bot");
    if(window.location.hash==="#chatbot"||window.location.hash==="#apex-chatbot")root.classList.add("is-open");
  }

  function addMessage(text,who,links){
    const log=root.querySelector(".apex-chatbot__messages");
    if(!log)return null;
    const message=document.createElement("div");
    message.className=`apex-chatbot__msg apex-chatbot__msg--${who}`;
    const body=document.createElement("p");
    body.textContent=text;
    message.appendChild(body);
    if(links&&links.length){
      const row=document.createElement("div");
      row.className="apex-chatbot__links";
      links.forEach((link)=>{
        const a=document.createElement("a");
        a.className="apex-chatbot__link";
        a.href=link.href;
        a.textContent=link.label;
        if(link.href.startsWith("mailto:"))a.setAttribute("target","_blank");
        row.appendChild(a);
      });
      message.appendChild(row);
    }
    log.appendChild(message);
    log.scrollTop=log.scrollHeight;
    return message;
  }

  function apiBase(){
    const explicit=root.dataset.apiUrl||window.MEDMASIA_BRIDGEBOT_API_URL||"";
    if(explicit)return explicit.replace(/\/+$/,"");
    return "";
  }

  async function askBridgeBot(question){
    const base=apiBase();
    if(!base)return null;
    const sessionId=sessionStorage.getItem(sessionKey)||"";
    const response=await fetch(`${base}/api/chat`,{
      method:"POST",
      headers:{"Content-Type":"application/json"},
      body:JSON.stringify({message:question,session_id:sessionId,language:"pt-BR",site:"medmasia"})
    });
    if(!response.ok)throw new Error(`BridgeBot ${response.status}`);
    const payload=await response.json();
    if(payload.session_id)sessionStorage.setItem(sessionKey,payload.session_id);
    return payload;
  }

  function addContextForm(){
    const base=apiBase();
    const sessionId=sessionStorage.getItem(sessionKey)||"";
    const log=root.querySelector(".apex-chatbot__messages");
    if(!base||!sessionId||!log||root.querySelector(".apex-chatbot__context"))return;
    const form=document.createElement("form");
    form.className="apex-chatbot__context apex-chatbot__msg apex-chatbot__msg--bot";
    form.innerHTML=`
      <input name="name" type="text" autocomplete="name" placeholder="Nome" required>
      <input name="clinic" type="text" autocomplete="organization" placeholder="Clinica ou especialidade">
      <input name="contact" type="text" autocomplete="email" placeholder="Email ou telefone" required>
      <textarea name="main_pain" rows="3" placeholder="Principal rotina, decisão ou gargalo" required></textarea>
      <label><input name="consent" type="checkbox" required> <span>Autorizo o contato para continuidade desta conversa.</span></label>
      <button type="submit">Enviar contexto</button>`;
    form.addEventListener("submit",async(event)=>{
      event.preventDefault();
      const data=new FormData(form);
      const contact=String(data.get("contact")||"").trim();
      const payload={
        session_id:sessionId,
        name:String(data.get("name")||"").trim(),
        company:String(data.get("clinic")||"").trim(),
        email:contact.includes("@")?contact:"",
        phone:contact.includes("@")?"":contact,
        main_pain:String(data.get("main_pain")||"").trim(),
        consent:data.get("consent")==="on",
        language:"pt-BR",
        site:"medmasia"
      };
      try{
        const response=await fetch(`${base}/api/lead`,{
          method:"POST",
          headers:{"Content-Type":"application/json"},
          body:JSON.stringify(payload)
        });
        if(!response.ok)throw new Error(`Registro ${response.status}`);
        form.remove();
        addMessage("Contexto registrado. A conversa foi marcada para revisão humana.","bot",[{label:"Contato",href:"#contact"}]);
      }catch(_){
        addMessage("Não consegui registrar agora. Você também pode usar o email de contato.","bot",[{label:"Email",href:`mailto:${email}`}]);
      }
    });
    log.appendChild(form);
    log.scrollTop=log.scrollHeight;
  }

  function findAnswer(question){
    const q=normalize(question);
    let best=null;
    let bestScore=0;
    copy.entries.forEach((entry)=>{
      const score=entry.keys.reduce((total,key)=>total+(q.includes(normalize(key))?1:0),0);
      if(score>bestScore){best=entry;bestScore=score;}
    });
    return best||{answer:copy.fallback,links:copy.fallbackLinks};
  }

  async function ask(question){
    const clean=String(question||"").trim();
    if(!clean)return;
    root.classList.add("is-open");
    addMessage(clean,"user");
    const loading=addMessage("Lendo o contexto disponível...","bot");
    try{
      const payload=await askBridgeBot(clean);
      if(payload&&payload.answer){
        loading.remove();
        const links=payload.collect_lead?[{label:"Solicitar conversa",href:"#contact"}]:[];
        addMessage(payload.answer,"bot",links);
        if(payload.collect_lead)addContextForm();
        return;
      }
    }catch(_){
      loading.remove();
    }
    const answer=findAnswer(clean);
    window.setTimeout(()=>addMessage(answer.answer,"bot",answer.links),160);
  }

  renderShell();
})();
