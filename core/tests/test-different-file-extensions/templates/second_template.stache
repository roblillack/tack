<html>
    <meta>
        <title>{{title}}</title>
        <style>
body {
    font-family: sans-serif;
    margin: 5rem 10%;
    line-height: 1.5;
}
nav>b {
    background-color: yellow;
}
.footer {
    font-size: 0.8rem;
    color: gray;
    border-top: 0.05rem solid gray;
    margin-top: 3rem;
}
        </style>
    </meta>

    <body>
        <nav>
        {{#navigation}}
        {{^current}}<a href="{{permalink}}">{{name}}</a>{{/current}}
        {{#current}}<b>{{name}}</b>{{/current}}
        {{/navigation}}
        </nav>

        {{{body}}}

        {{#products}}
        <h2>{{name}}</h2>
        {{desc}}
        {{/products}}

        <div class="footer">
            Siblings of this page are:
            {{#siblings}}
            {{name}}
            {{/siblings}}
        </div>

    </body>
</html>