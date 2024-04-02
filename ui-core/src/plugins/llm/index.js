import olama from "./olama";

function createPlugin () {
    /**
     * Start the plugin.
     *
     * @param {object} params The start parameters.
     * @param {object} params.config The configuration options.
     * @param {object} params.eventBus The cross-plugin event emitter.
     * @param {object} params.plugins All other plugins.
     * @returns {{api:object,events:string[],name:string}} The plugin API.
     */
    function start ({ config, eventBus, plugins }) {
      // debug.enabled = config.debug;
  
      const {  } = config;
      const {  } = plugins;
    
      // eventBus.on('coin-block', emitLumerinStatus);
  
      // Collect meta parsers
      const metaParsers = Object.assign({},
        // {
        //   // auction: auctionEvents.auctionMetaParser,
        //   // converter: converterEvents.converterMetaParser,
        //   export: porterEvents.exportMetaParser,
        //   import: porterEvents.importMetaParser,
        //   importRequest: porterEvents.importRequestMetaParser
        // },
      );
  
      // Build and return API
      return {
        api: {
          ...olama
        },
        events: [
        ],
        name: 'llm'
      };
    }
  
    /**
     * Stop the plugin.
     */
    function stop () {}
  
    return {
      start,
      stop
    };
  }