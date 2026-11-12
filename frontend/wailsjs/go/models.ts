export namespace main {
	
	export class Option {
	    key: string;
	    summary: string;
	    doc: string;
	    summaryJa?: string;
	    docJa?: string;
	    default: string;
	    kind: string;
	    choices: string[];
	    category: string;
	    reload: string;
	
	    static createFrom(source: any = {}) {
	        return new Option(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.summary = source["summary"];
	        this.doc = source["doc"];
	        this.summaryJa = source["summaryJa"];
	        this.docJa = source["docJa"];
	        this.default = source["default"];
	        this.kind = source["kind"];
	        this.choices = source["choices"];
	        this.category = source["category"];
	        this.reload = source["reload"];
	    }
	}
	export class Result {
	    ok: boolean;
	    message: string;
	    reloaded: boolean;
	    set: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	        this.reloaded = source["reloaded"];
	        this.set = source["set"];
	    }
	}
	export class State {
	    version: string;
	    configPath: string;
	    backupPath: string;
	    options: Option[];
	    set: Record<string, string>;
	    lang: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.configPath = source["configPath"];
	        this.backupPath = source["backupPath"];
	        this.options = this.convertValues(source["options"], Option);
	        this.set = source["set"];
	        this.lang = source["lang"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Theme {
	    name: string;
	    background: string;
	    foreground: string;
	    palette: string[];
	    dark: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Theme(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.background = source["background"];
	        this.foreground = source["foreground"];
	        this.palette = source["palette"];
	        this.dark = source["dark"];
	    }
	}

}

